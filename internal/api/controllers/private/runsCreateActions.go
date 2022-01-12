package private

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"

	"playbook-dispatcher/internal/api/connectors"
	"playbook-dispatcher/internal/api/controllers/public"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/api/middleware"
	dbModel "playbook-dispatcher/internal/common/model/db"

	"github.com/google/uuid"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	ansibleCloudConnectorDirective = "rhc-worker-playbook"
	satCloudConnectorDirective     = "playbook-sat"
)

type RunInputGeneric struct {
	Recipient       string
	Account         string
	Url             string
	Hosts           *RunInputHosts
	Labels          *public.Labels
	Timeout         *public.RunTimeout
	OrgId           *string
	RecipientConfig *RecipientConfig
	Name            *string
	WebConsoleUrl   *string
	Principal       *string
	CorrelationId   uuid.UUID
	Directive       string
}

func getCorrelationId(cfg *viper.Viper) uuid.UUID {
	if cfg.GetBool("demo.mode") {
		return uuid.UUID{}
	}
	return uuid.New()
}

func RunInputV1GenericMap(runInput RunInput, cfg *viper.Viper) RunInputGeneric {
	return RunInputGeneric{
		Recipient:     string(runInput.Recipient),
		Account:       string(runInput.Account),
		Url:           string(runInput.Url),
		Labels:        runInput.Labels,
		Timeout:       runInput.Timeout,
		Hosts:         runInput.Hosts,
		CorrelationId: getCorrelationId(cfg),
		Directive:     ansibleCloudConnectorDirective,
	}
}

func RunInputV2GenericMap(runInput RunInputV2, satReq bool, cfg *viper.Viper) RunInputGeneric {
	orgIdString := string(runInput.OrgId)
	playbookName := string(runInput.Name)
	principal := string(runInput.Principal)

	webConsoleUrl := cfg.GetString("web.console.url.default")
	if runInput.WebConsoleUrl != nil {
		webConsoleUrl = string(*runInput.WebConsoleUrl)
	}

	directive := ansibleCloudConnectorDirective
	if satReq {
		directive = satCloudConnectorDirective
	}
	return RunInputGeneric{
		Recipient:       string(runInput.Recipient),
		OrgId:           &orgIdString,
		Url:             string(runInput.Url),
		Labels:          runInput.Labels,
		Timeout:         runInput.Timeout,
		Hosts:           runInput.Hosts,
		Name:            &playbookName,
		WebConsoleUrl:   &webConsoleUrl,
		Principal:       &principal,
		RecipientConfig: runInput.RecipientConfig,
		CorrelationId:   getCorrelationId(cfg),
		Directive:       directive,
	}
}

func CheckV2ReqFields(runInput RunInputV2) (ansible bool, satellite bool) {
	ansible, satellite = true, true

	if runInput.RecipientConfig == nil {
		satellite = false
	}
	if runInput.RecipientConfig != nil && (runInput.RecipientConfig.SatId == nil || runInput.RecipientConfig.SatOrgId == nil) {
		satellite = false
	}

	if runInput.Hosts == nil {
		satellite = false
	} else {
		for _, host := range *runInput.Hosts {
			if host.AnsibleHost == nil || *host.AnsibleHost == "" {
				ansible = false
			}
			if host.InventoryId == nil {
				satellite = false
			}
		}
	}
	return
}

func BuildCloudConnectorMetadata(satRequest bool, runInput RunInputGeneric) (metadata map[string]string) {
	if !satRequest {
		metadata = map[string]string{
			"crc_dispatcher_correlation_id": runInput.CorrelationId.String(),
		}
		return
	}

	var hosts string
	for _, host := range *runInput.Hosts {
		hosts += *host.InventoryId
	}

	principal := *runInput.Principal
	var principalHash string

	if principal != "" {
		principalHash = fmt.Sprintf("%x", sha256.Sum256([]byte(principal)))
	}

	metadata = map[string]string{
		"operation":         "run",
		"correlation_id":    runInput.CorrelationId.String(),
		"playbook_run_name": *runInput.Name,
		"playbook_run_url":  string(*runInput.WebConsoleUrl),
		"sat_id":            string(*runInput.RecipientConfig.SatId),
		"sat_org_id":        string(*runInput.RecipientConfig.SatOrgId),
		"initiator_user_id": principalHash,
		"hosts":             hosts,
	}
	return
}

func sendToCloudConnector(
	satReq bool,
	account string,
	recipient uuid.UUID,
	runInput RunInputGeneric,
	cloudConnectorClient connectors.CloudConnectorClient,
	context context.Context,
) *RunCreated {
	cloudConnectorMetadata := BuildCloudConnectorMetadata(satReq, runInput)

	messageId, notFound, err := cloudConnectorClient.SendCloudConnectorRequest(
		context,
		account,
		recipient,
		runInput.Url,
		runInput.Directive,
		cloudConnectorMetadata,
	)

	requestType := "ansible"
	if satReq {
		requestType = "satellite"
	}

	if err != nil {
		instrumentation.CloudConnectorRequestError(context, err, recipient, requestType)
		return runCreateError(http.StatusInternalServerError)
	} else if notFound {
		instrumentation.CloudConnectorNoConnection(context, recipient, requestType)
		return runCreateError(http.StatusNotFound)
	}

	instrumentation.CloudConnectorOK(context, recipient, messageId)
	return nil
}

func newRun(input *RunInputGeneric, correlationId uuid.UUID, status string, recipient uuid.UUID, v2Req bool, satReq bool, cfg *viper.Viper) dbModel.Run {
	run := dbModel.Run{
		ID:            uuid.New(),
		Account:       input.Account,
		CorrelationID: correlationId,
		URL:           input.Url,
		Status:        status,
		Recipient:     recipient,
	}

	if input.Labels != nil {
		run.Labels = input.Labels.AdditionalProperties
	}

	if input.Timeout != nil {
		run.Timeout = int(*input.Timeout)
	} else {
		run.Timeout = cfg.GetInt("default.run.timeout")
	}

	if v2Req {
		run.OrgID = *input.OrgId
		run.PlaybookName = *input.Name
		run.Principal = *input.Principal
		run.PlaybookRunUrl = *input.WebConsoleUrl

		run.ResponseFull = cfg.GetBool("satellite.response.full")

	}

	if satReq {
		satUUID, _ := uuid.Parse(string(*input.RecipientConfig.SatId))

		run.SatId = satUUID
		run.SatOrgId = string(*input.RecipientConfig.SatOrgId)
	}

	return run
}

func newHostRun(runHosts *RunInputHosts, entityId uuid.UUID, satReq bool) []dbModel.RunHost {
	newHosts := make([]dbModel.RunHost, len(*runHosts))

	for i, inputHost := range *runHosts {
		newHosts[i] = dbModel.RunHost{
			ID:          uuid.New(),
			RunID:       entityId,
			InventoryID: nil,
			Status:      dbModel.RunStatusRunning,
		}

		if !satReq {
			newHosts[i].Host = string(*inputHost.AnsibleHost)
		}

		if inputHost.InventoryId != nil {
			inventoryID := uuid.MustParse(string(*inputHost.InventoryId))
			newHosts[i].InventoryID = &inventoryID

			if satReq {
				newHosts[i].Host = string(*inputHost.InventoryId)
			}
		}
	}

	return newHosts
}

func recordRunInformation(
	runInput RunInputGeneric,
	recipient uuid.UUID,
	correlationId uuid.UUID,
	v2Req bool,
	satReq bool,
	requestType string,
	database *gorm.DB,
	cfg *viper.Viper,
	context context.Context,
) (*public.RunId, *RunCreated) {
	entity := newRun(&runInput, correlationId, dbModel.RunStatusRunning, recipient, v2Req, satReq, cfg)
	entity.Service = middleware.GetPSKPrincipal(context)

	if dbResult := database.Create(&entity); dbResult.Error != nil {
		instrumentation.PlaybookRunCreateError(context, dbResult.Error, &entity, requestType)
		return nil, runCreateError(http.StatusInternalServerError)
	}

	if runInput.Hosts != nil {
		newHosts := newHostRun(runInput.Hosts, entity.ID, satReq)

		if dbResult := database.Create(newHosts); dbResult.Error != nil {
			instrumentation.PlaybookRunHostCreateError(context, dbResult.Error, newHosts, requestType)
			return nil, runCreateError(http.StatusInternalServerError)
		}
	}

	runId := public.RunId(entity.ID.String())
	instrumentation.RunCreated(context, recipient, entity.ID, runInput.Url, entity.Service, requestType)

	return &runId, nil
}

func runCreateError(code int) *RunCreated {
	return &RunCreated{
		Code: code,
		// TODO report error
	}
}
