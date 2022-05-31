package public

import (
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/api/middleware"
	"playbook-dispatcher/internal/api/rbac"
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils"

	"github.com/google/uuid"
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
		WithContext(ctx.Request().Context()).
		Table("run_hosts").
		Joins("INNER JOIN runs on runs.id = run_hosts.run_id").
		Where("runs.account = ?", identity.Identity.AccountNumber)

	permissions := middleware.GetPermissions(ctx)
	if allowedServices := rbac.GetPredicateValues(permissions, "service"); len(allowedServices) > 0 {
		queryBuilder.Where("runs.service IN ?", allowedServices)
	}

	if params.Filter != nil {
		if params.Filter.Status != nil {
			status := *params.Filter.Status
			switch status {
			case dbModel.RunStatusTimeout:
				queryBuilder.Where("runs.status = 'timeout' OR runs.status = 'running' AND runs.created_at + runs.timeout * interval '1 second' <= NOW()")
			case dbModel.RunStatusRunning:
				queryBuilder.Where("run_hosts.status = ?", status)
				queryBuilder.Where("runs.created_at + runs.timeout * interval '1 second' > NOW()")
			default:
				queryBuilder.Where("run_hosts.status = ?", status)
			}
		}

		if runFilters := middleware.GetDeepObject(ctx, "filter", "run"); len(runFilters) > 0 {
			if id, ok := runFilters["id"]; ok {
				queryBuilder.Where("run_hosts.run_id = ?", id)
			}

			if service, ok := runFilters["service"]; ok {
				queryBuilder.Where("runs.service = ?", service)
			}
		}

		if labelFilters := middleware.GetDeepObject(ctx, "filter", "run", "labels"); len(labelFilters) > 0 {
			for key, values := range labelFilters {
				for _, value := range values {
					queryBuilder.Where("runs.labels ->> ? = ?", key, value)
				}
			}
		}

		if params.Filter.InventoryId != nil {
			parsedInventoryID, err := uuid.Parse(string(*params.Filter.InventoryId))
			if err != nil {
				return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("invalid inventory_id: %s", *params.Filter.InventoryId))
			}

			queryBuilder.Where("run_hosts.inventory_id = ?", parsedInventoryID)
		}
	}

	var total int64
	countResult := queryBuilder.Count(&total)

	if countResult.Error != nil {
		instrumentation.PlaybookRunReadError(ctx, countResult.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	queryBuilder.Limit(limit)
	queryBuilder.Offset(offset)

	queryBuilder.Select(utils.MapStrings(fields, mapHostFieldsToSql))

	var dbRunHosts []dbModel.RunHost
	dbResult := queryBuilder.Find(&dbRunHosts)

	if dbResult.Error != nil {
		instrumentation.PlaybookRunReadError(ctx, dbResult.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	hosts := []RunHost{}

	for _, host := range dbRunHosts {

		runHost := RunHost{}
		runId := RunId(host.RunID.String())
		runStatus := RunStatus(host.Status)

		for _, field := range fields {
			switch field {
			case fieldHost:
				runHost.Host = utils.StringRef(host.Host)
			case fieldStdout:
				runHost.Stdout = utils.StringRef(host.Log)
			case fieldStatus:
				runHost.Status = &runStatus
			case fieldRun:
				runHost.Run = &Run{
					Id: &runId,
				}
			case fieldLinks:
				runHost.Links = &RunHostLinks{
					InventoryHost: inventoryLink(host.InventoryID),
				}
			case fieldInventoryId:
				if host.InventoryID != nil {
					inventoryID := InventoryId(host.InventoryID.String())
					runHost.InventoryId = &inventoryID
				}
			}
		}

		hosts = append(hosts, runHost)
	}

	return ctx.JSON(http.StatusOK, &RunHosts{
		Data: hosts,
		Meta: Meta{
			Count: len(hosts),
			Total: int(total),
		},
		Links: createLinks("/api/playbook-dispatcher/v1/run_hosts", middleware.GetQueryString(ctx), getLimit(params.Limit), getOffset(params.Offset), int(total)),
	})
}

func mapHostFieldsToSql(field string) string {
	switch field {
	case "host":
		return "run_hosts.host"
	case "run":
		return "run_hosts.run_id"
	case "status":
		return "run_hosts.status"
	case "stdout":
		return "run_hosts.log"
	case fieldLinks:
		return "run_hosts.inventory_id"
	case fieldInventoryId:
		return "run_hosts.inventory_id"
	default:
		panic("unknown field " + field)
	}
}

func inventoryLink(inventoryID *uuid.UUID) *string {
	if inventoryID == nil {
		return nil
	}

	link := fmt.Sprintf("/api/inventory/v1/hosts/%s", inventoryID.String())
	return &link
}
