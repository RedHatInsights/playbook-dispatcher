package responseConsumer

import (
	"context"
	"errors"
	"time"

	"playbook-dispatcher/internal/common/ansible"
	"playbook-dispatcher/internal/common/constants"
	kafkaUtils "playbook-dispatcher/internal/common/kafka"
	"playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/satellite"
	"playbook-dispatcher/internal/common/utils"
	"playbook-dispatcher/internal/response-consumer/instrumentation"

	k "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/google/uuid"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	EventPlaybookOnStats  = "playbook_on_stats"
	EventRunnerOnFailed   = "runner_on_failed"
	EventExecutorOnFailed = "executor_on_failed"

	EventSatPlaybookFinished  = "playbook_run_finished"
	EventSatPlaybookCompleted = "playbook_run_completed"

	EventSatStatusFailure  = "failure"
	EventSatStatusSuccess  = "success"
	EventSatStatusCanceled = "canceled"
)

type handler struct {
	db *gorm.DB
}

func (this *handler) BeforeUpdate(ctx context.Context, tx *gorm.DB) (err error) {
	if !tx.Statement.Changed("SatSequence") {
		instrumentation.PlaybookRunUpdateSequenceOrder(ctx)
	}

	return nil
}

func (this *handler) onMessage(ctx context.Context, msg *k.Message) {
	requestId, correlationId, requestType, err := getHeaders(msg)

	if err != nil {
		instrumentation.CannotReadHeaders(ctx, err)
		return
	}

	ctx = utils.WithRequestId(ctx, requestId)
	ctx = utils.WithCorrelationId(ctx, correlationId.String())

	value := parseMessage(ctx, requestType, msg)
	if value == nil {
		return
	}

	ctx = utils.WithOrgId(ctx, value.OrgId)

	utils.GetLogFromContext(ctx).Debugw("Processing message",
		"upload_timestamp", value.UploadTimestamp,
		"topic", *msg.TopicPartition.Topic,
		"partition", msg.TopicPartition.Partition,
		"offset", msg.TopicPartition.Offset.String(),
	)

	var status string
	var eventsSerialized []byte

	var runsUpdated int64

	run := db.Run{}

	err = this.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		baseQuery := tx.Model(db.Run{}).
			Where("org_id = ?", value.OrgId).
			Where("correlation_id = ?", correlationId)

		selectResult := baseQuery.Select("id", "status", "response_full").First(&run)

		if requestType == satMessageHeaderValue {
			satellite.SortSatEvents(value.SatEvents)

			status = inferSatPlaybookStatus(value.SatEvents)
			eventsSerialized = utils.MustMarshal(value.SatEvents)

			if !run.ResponseFull {
				status = checkSatStatusPartial(value.SatEvents)
			}

			if run.Status == db.RunStatusFailure || run.Status == db.RunStatusCanceled {
				status = run.Status
			}
		} else {
			status = inferStatus(value.RunnerEvents, nil)
			eventsSerialized = utils.MustMarshal(value.RunnerEvents)
		}

		if selectResult.Error != nil {
			if errors.Is(selectResult.Error, gorm.ErrRecordNotFound) {
				return nil
			}

			utils.GetLogFromContext(ctx).Errorw("Error fetching run from db", "error", selectResult.Error)
			return selectResult.Error
		}

		toUpdate := db.Run{
			Status: status,
			Events: eventsSerialized,
		}

		// We retrieved the run id (primary key) above...lets use it here during the update to speed up the query
		baseQuery = baseQuery.Where("id = ?", run.ID)

		// Only update if the run is not marked as complete
		updateResult := baseQuery.Where("status not in ?", []string{db.RunStatusSuccess, db.RunStatusFailure}).Select("status", "events").Updates(toUpdate)
		if updateResult.Error != nil {
			utils.GetLogFromContext(ctx).Errorw("Error updating run in db", "error", updateResult.Error)
			return updateResult.Error
		} else {
			runsUpdated = updateResult.RowsAffected
		}

		var toCreate []db.RunHost

		if requestType == runnerMessageHeaderValue {
			hosts := ansible.GetAnsibleHosts(*value.RunnerEvents)

			if len(hosts) == 0 {
				// If the the playbook fials the signature validation step or if ansible is not
				// installed, then the generated output will not have any events with a "host" field.
				// When this happens (the hosts list is empty), then we need to add a "localhost"
				// entry to the hosts list so that output from the run will get inserted into the
				// host table otherwise the output gets thrown away.
				utils.GetLogFromContext(ctx).Debug("Unable to locate any hosts in the ansible output...setting hosts to [localhost]")
				hosts = []string{"localhost"}
			}

			toCreate = mapHostsToRunHosts(hosts, func(host string) db.RunHost {
				return db.RunHost{
					ID:     uuid.New(),
					RunID:  run.ID,
					Host:   host,
					Status: inferStatus(value.RunnerEvents, &host),
					Log:    ansible.GetStdout(*value.RunnerEvents, nil),
				}
			})
			return createRecord(ctx, tx, toCreate)
		} else if requestType == satMessageHeaderValue {
			hosts := satellite.GetSatHosts(*value.SatEvents)

			if len(hosts) == 0 {
				return nil
			}

			toCreate = mapHostsToRunHosts(hosts, func(host string) db.RunHost {
				satHost := satellite.GetSatHostInfo(*value.SatEvents, &host)
				inventoryId := uuid.MustParse(host)
				return db.RunHost{
					ID:          uuid.New(),
					RunID:       run.ID,
					InventoryID: &inventoryId,
					SatSequence: satHost.Sequence,
					Status:      inferSatHostStatus(value.SatEvents, host),
					Log:         satHost.Console,
				}
			})
			return satUpdateRecord(ctx, tx, run.ResponseFull, toCreate)
		}

		return nil
	})

	if err != nil {
		instrumentation.PlaybookRunUpdateError(ctx, err, status, run.ID)
	} else if runsUpdated > 0 {
		instrumentation.PlaybookRunUpdated(ctx, status, run.ID)
	} else {
		instrumentation.PlaybookRunUpdateMiss(ctx, status)
	}
}

func satAssignmentWithCase(responseFull bool, updateHost db.RunHost) map[string]interface{} {
	satSequence, status, log := *updateHost.SatSequence, updateHost.Status, updateHost.Log

	updateMap := map[string]interface{}{
		"status":       status,
		"sat_sequence": satSequence,
		"log":          log,
	}

	if !responseFull {
		updateMap["log"] = gorm.Expr(`CASE WHEN (sat_sequence IS NULL AND ? > 0) OR sat_sequence + 1 < ? THEN log || '\n\u2026\n' || ? ELSE log || ? END`, satSequence, satSequence, log, log)
	}

	return updateMap
}

func satUpdateRecord(ctx context.Context, tx *gorm.DB, responseFull bool, toUpdate []db.RunHost) error {
	for _, runHost := range toUpdate {
		resultValues := db.RunHost{}
		updateResult := tx.Model(&resultValues)

		if runHost.SatSequence != nil {
			updateResult.Clauses(clause.Returning{}).Where("run_id = ? AND inventory_id = ? AND (sat_sequence IS NULL OR sat_sequence < ?)", runHost.RunID, runHost.InventoryID, *runHost.SatSequence).
				Updates(satAssignmentWithCase(responseFull, runHost))
		} else {
			// only update status when runHost.SatSequence is nil e.g. when runHost finished
			updateResult.Where("run_id = ? AND inventory_id = ?", runHost.RunID, runHost.InventoryID).
				Updates(map[string]interface{}{"status": runHost.Status})
		}

		if updateResult.Error != nil {
			utils.GetLogFromContext(ctx).Errorw("Error updating satellite host in db", "error", updateResult.Error)
			return updateResult.Error
		}
	}
	return nil
}

func createRecord(ctx context.Context, tx *gorm.DB, toCreate []db.RunHost) error {

	successOrFailure := clause.OrConditions{Exprs: []clause.Expression{
		clause.Eq{Column: "run_hosts.status", Value: db.RunStatusSuccess},
		clause.Eq{Column: "run_hosts.status", Value: db.RunStatusFailure},
	}}

	notMarkedAsComplete := clause.Where{Exprs: []clause.Expression{clause.Not(successOrFailure)}}

	createResult := tx.Model(db.RunHost{}).
		Clauses(clause.OnConflict{
			Where:     notMarkedAsComplete,
			Columns:   []clause.Column{{Name: "run_id"}, {Name: "host"}},
			DoUpdates: clause.AssignmentColumns([]string{"status", "log"}),
		}).
		Create(&toCreate)

	if createResult.Error != nil {
		utils.GetLogFromContext(ctx).Errorw("Error upserting run hosts in db", "error", createResult.Error)
		return createResult.Error
	}

	return nil
}

func inferStatus(events *[]message.PlaybookRunResponseMessageYamlEventsElem, host *string) string {
	finished := false
	failed := false

	for _, event := range *events {
		if event.Event == EventPlaybookOnStats {
			finished = true
		}

		// if host parameter is defined only consider events for the given host
		if host != nil && event.EventData != nil && event.EventData.Host != nil && *event.EventData.Host != *host {
			continue
		}

		if event.Event == EventRunnerOnFailed {
			failed = true
		}

		if event.Event == EventExecutorOnFailed {
			failed = true
			finished = true
		}
	}

	switch {
	case finished && failed:
		return db.RunStatusFailure
	case finished && !failed:
		return db.RunStatusSuccess
	default:
		return db.RunStatusRunning
	}
}

func satStatusEventDbMap(status message.PlaybookSatRunResponseMessageYamlEventsElemStatus) string {
	switch {
	case status == EventSatStatusSuccess:
		return db.RunStatusSuccess
	case status == EventSatStatusFailure:
		return db.RunStatusFailure
	case status == EventSatStatusCanceled:
		return db.RunStatusCanceled
	default:
		return db.RunStatusRunning
	}
}

func inferSatPlaybookStatus(events *[]message.PlaybookSatRunResponseMessageYamlEventsElem) string {
	hostStatusMap := make(map[string]string)

	for _, event := range *events {
		if event.Type == EventSatPlaybookCompleted {
			return satStatusEventDbMap(*event.Status)
		}

		if event.Host != nil {
			if _, ok := hostStatusMap[*event.Host]; !ok {
				hostStatusMap[*event.Host] = db.RunStatusRunning
			}
			if event.Status != nil {
				hostStatusMap[*event.Host] = satStatusEventDbMap(*event.Status)
			}
		}
	}

	failed := false
	canceled := false

	for _, status := range hostStatusMap {
		if status == db.RunStatusRunning {
			return status
		}
		if status == db.RunStatusFailure {
			failed = true
		}
		if status == db.RunStatusCanceled {
			canceled = true
		}
	}

	switch {
	case failed:
		return db.RunStatusFailure
	case canceled:
		return db.RunStatusCanceled
	default:
		return db.RunStatusSuccess
	}
}

func inferSatHostStatus(events *[]message.PlaybookSatRunResponseMessageYamlEventsElem, host string) string {
	for _, event := range *events {
		if event.Host != nil && *event.Host != host {
			continue
		}
		if event.Type == EventSatPlaybookFinished && event.Status != nil {
			return satStatusEventDbMap(*event.Status)
		}
	}

	return db.RunStatusRunning
}

func checkSatStatusPartial(events *[]message.PlaybookSatRunResponseMessageYamlEventsElem) string {
	// for response_full = false, set run status to "running" unless "playbook_run_completed" signal is received
	for _, event := range *events {
		if event.Type != EventSatPlaybookCompleted || event.Status == nil {
			continue
		}

		return satStatusEventDbMap(*event.Status)
	}

	return db.RunStatusRunning
}

type parsedMessageInfo struct {
	OrgId           string
	B64Identity     string
	UploadTimestamp string
	RunnerEvents    *[]message.PlaybookRunResponseMessageYamlEventsElem
	SatEvents       *[]message.PlaybookSatRunResponseMessageYamlEventsElem
}

func parseMessage(ctx context.Context, requestType string, msg *k.Message) *parsedMessageInfo {
	if requestType == runnerMessageHeaderValue {
		value := &message.PlaybookRunResponseMessageYaml{}

		if err := value.UnmarshalJSON(msg.Value); err != nil {
			instrumentation.UnmarshallIncomingMessageError(ctx, err)
			return nil
		}

		return &parsedMessageInfo{
			OrgId:           value.OrgId,
			B64Identity:     value.B64Identity,
			UploadTimestamp: value.UploadTimestamp.Format(time.RFC3339),
			RunnerEvents:    &value.Events,
		}
	} else {
		value := &message.PlaybookSatRunResponseMessageYaml{}

		if err := value.UnmarshalJSON(msg.Value); err != nil {
			instrumentation.UnmarshallIncomingMessageError(ctx, err)
			return nil
		}

		return &parsedMessageInfo{
			OrgId:           value.OrgId,
			B64Identity:     value.B64Identity,
			UploadTimestamp: value.UploadTimestamp.Format(time.RFC3339),
			SatEvents:       &value.Events,
		}
	}
}

func getHeaders(msg *k.Message) (requestId string, correlationId uuid.UUID, requestType string, err error) {
	if requestId, err = kafkaUtils.GetHeader(msg, constants.HeaderRequestId); err != nil {
		return
	}

	var correlationIdRaw string
	if correlationIdRaw, err = kafkaUtils.GetHeader(msg, constants.HeaderCorrelationId); err != nil {
		return
	}

	if correlationId, err = uuid.Parse(correlationIdRaw); err != nil {
		return
	}

	if requestType, err = kafkaUtils.GetHeader(msg, constants.HeaderRequestType); err != nil {
		return
	}

	return
}

func mapHostsToRunHosts(hosts []string, fn func(host string) db.RunHost) []db.RunHost {
	result := make([]db.RunHost, len(hosts))
	for i, host := range hosts {
		result[i] = fn(host)
	}

	return result
}
