package controllers

import (
	"net/http"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/api/middleware"
	"playbook-dispatcher/internal/common/ansible"
	dbModel "playbook-dispatcher/internal/common/model/db"
	messageModel "playbook-dispatcher/internal/common/model/message"
	"playbook-dispatcher/internal/common/utils"

	"github.com/labstack/echo/v4"
	identityMiddleware "github.com/redhatinsights/platform-go-middlewares/identity"
)

func (this *controllers) ApiRunHostsList(ctx echo.Context, params ApiRunHostsListParams) error {
	identity := identityMiddleware.Get(ctx.Request().Context())

	limit := getLimit(params.Limit)
	offset := getOffset(params.Offset)

	fields, err := parseFields(middleware.GetDeepObject(ctx, "fields"), "data", runHostFields, defaultRunHostFields)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	queryBuilder := this.database.
		Select(
			"id",
			"events",
			`CASE WHEN runs.status='running' AND runs.created_at + runs.timeout * interval '1 second' <= NOW() THEN 'timeout' ELSE runs.status END as status`,
		).
		Where("account = ?", identity.Identity.AccountNumber).
		Order("created_at desc").
		Order("id")

	if params.Filter != nil {
		if params.Filter.Status != nil { // TODO: possible 1-n mapping between runs and hosts
			status := *params.Filter.Status
			switch status {
			case dbModel.RunStatusTimeout:
				queryBuilder.Where("runs.created_at + runs.timeout * interval '1 second' <= NOW()")
				status = dbModel.RunStatusRunning
			case dbModel.RunStatusRunning:
				queryBuilder.Where("runs.created_at + runs.timeout * interval '1 second' > NOW()")
			}

			queryBuilder.Where("runs.status = ?", status)
		}

		if runFilters := middleware.GetDeepObject(ctx, "filter", "run"); len(runFilters) > 0 {
			if id, ok := runFilters["id"]; ok {
				queryBuilder.Where("runs.id = ?", id)
			}
		}

		if labelFilters := middleware.GetDeepObject(ctx, "filter", "run", "labels"); len(labelFilters) > 0 {
			for key, values := range labelFilters {
				for _, value := range values {
					queryBuilder.Where("runs.labels ->> ? = ?", key, value)
				}
			}
		}
	}

	var dbRuns []dbModel.Run
	dbResult := queryBuilder.Find(&dbRuns)

	if dbResult.Error != nil {
		instrumentation.PlaybookRunReadError(ctx, dbResult.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	hosts := []RunHost{}

	var events []messageModel.PlaybookRunResponseMessageYamlEventsElem

	for _, run := range dbRuns {
		utils.MustUnmarshal(run.Events, &events)

		runId := RunId(run.ID.String())
		runStatus := RunStatus(run.Status)

		for _, host := range ansible.GetAnsibleHosts(events) {
			stdout := ansible.GetStdout(events)
			runHost := RunHost{}

			for _, field := range fields {
				switch field {
				case fieldHost:
					runHost.Host = &host
				case fieldStdout:
					runHost.Stdout = &stdout
				case fieldStatus:
					runHost.Status = &runStatus
				case fieldRun:
					runHost.Run = &Run{
						Id: &runId,
					}
				}
			}

			hosts = append(hosts, runHost)
		}

		// TODO: should be in the inner loop
		if len(hosts) == limit+offset {
			break
		}
	}

	// TODO: this is a very poor way of implementing pagination
	// it will be fixed with db model refactoring
	if offset > len(hosts) {
		hosts = hosts[0:0]
	} else {
		hosts = hosts[offset:utils.Min(limit+offset, len(hosts))]
	}

	return ctx.JSON(http.StatusOK, &RunHosts{
		Data: hosts,
		Meta: Meta{
			Count: len(hosts),
		},
	})
}
