package public

import (
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/api/middleware"
	"playbook-dispatcher/internal/api/rbac"
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils"
	"strings"

	"github.com/labstack/echo/v4"
	identityMiddleware "github.com/redhatinsights/platform-go-middlewares/identity"
)

func getOrderBy(params ApiRunsListParams) string {
	if params.SortBy == nil || len(*params.SortBy) == 0 {
		return "created_at desc"
	}

	if parts := strings.Split(string(*params.SortBy), ":"); len(parts) == 1 {
		return fmt.Sprintf("%s %s", parts[0], "desc")
	} else {
		return fmt.Sprintf("%s %s", parts[0], parts[1])
	}
}

func mapFieldsToSql(field string) string {
	// set status to "timeout" on read if the run has expired
	if field == fieldStatus {
		return `CASE WHEN runs.status='running' AND runs.created_at + runs.timeout * interval '1 second' <= NOW() THEN 'timeout' ELSE runs.status END as status`
	}

	// these column names for these fields are different in the db
	if field == fieldName {
		return "playbook_name"
	}

	if field == fieldWebConsoleUrl {
		return "playbook_run_url"
	}

	return field
}

func (this *controllers) ApiRunsList(ctx echo.Context, params ApiRunsListParams) error {
	var dbRuns []dbModel.Run

	identity := identityMiddleware.Get(ctx.Request().Context())
	db := this.database.WithContext(ctx.Request().Context())

	// tenant isolation
	queryBuilder := db.Table("runs").Where("account = ?", identity.Identity.AccountNumber)

	// rbac
	permissions := middleware.GetPermissions(ctx)
	if allowedServices := rbac.GetPredicateValues(permissions, "service"); len(allowedServices) > 0 {
		queryBuilder.Where("service IN ?", allowedServices)
	}

	fields, err := parseFields(middleware.GetDeepObject(ctx, "fields"), "data", runFields, defaultRunFields)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if params.Filter != nil {
		if params.Filter.Status != nil {
			status := *params.Filter.Status
			switch status {
			case dbModel.RunStatusTimeout:
				queryBuilder.Where("runs.status = 'timeout' OR runs.status = 'running' AND runs.created_at + runs.timeout * interval '1 second' <= NOW()")
			case dbModel.RunStatusRunning:
				queryBuilder.Where("runs.status = ?", status)
				queryBuilder.Where("runs.created_at + runs.timeout * interval '1 second' > NOW()")
			default:
				queryBuilder.Where("runs.status = ?", status)
			}
		}

		if params.Filter.Recipient != nil {
			queryBuilder.Where("runs.recipient = ?", *params.Filter.Recipient)
		}

		if params.Filter.Service != nil {
			queryBuilder.Where("runs.service = ?", *params.Filter.Service)
		}
	}

	if labelFilters := middleware.GetDeepObject(ctx, "filter", "labels"); len(labelFilters) > 0 {
		for key, values := range labelFilters {
			for _, value := range values {
				queryBuilder.Where("runs.labels ->> ? = ?", key, value)
			}
		}
	}

	var total int64
	countResult := queryBuilder.Count(&total)

	if countResult.Error != nil {
		instrumentation.PlaybookRunReadError(ctx, countResult.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	queryBuilder.Select(utils.MapStrings(fields, mapFieldsToSql))

	queryBuilder.Order(getOrderBy(params))
	queryBuilder.Order("id") // secondary criteria to guarantee stable sorting

	queryBuilder.Limit(getLimit(params.Limit))
	queryBuilder.Offset(getOffset(params.Offset))

	dbResult := queryBuilder.Find(&dbRuns)

	if dbResult.Error != nil {
		instrumentation.PlaybookRunReadError(ctx, dbResult.Error)
		return ctx.NoContent(http.StatusInternalServerError)
	}

	response := make([]Run, len(dbRuns))

	for i, v := range dbRuns {
		response[i] = *dbRuntoApiRun(&v, fields)
	}

	return ctx.JSON(http.StatusOK, &Runs{
		Data: response,
		Meta: Meta{
			Count: len(response),
			Total: int(total),
		},
		Links: createLinks("/api/playbook-dispatcher/v1/runs", middleware.GetQueryString(ctx), getLimit(params.Limit), getOffset(params.Offset), int(total)),
	})
}
