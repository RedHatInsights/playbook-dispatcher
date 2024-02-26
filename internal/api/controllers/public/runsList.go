package public

import (
	"encoding/json"
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
	"gorm.io/gorm"
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

	// column names for these fields are different in the db
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
	queryBuilder := db.Table("runs").Where("org_id = ?", identity.Identity.OrgID)

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
		queryBuilder, _ = addLabelFilterToQueryAsWhereClause(queryBuilder, labelFilters)
		// FIXME:  Don't eat the error!
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

func addLabelFilterToQueryAsWhereClause(queryBuilder *gorm.DB, labelFilters map[string][]string) (*gorm.DB, error) {
	labels := make(map[string]string)

	for key, values := range labelFilters {
		// The inner for loop seems kind of odd.  The labels are basically a
		// hash map. As a result, you cannot have duplicate keys.  However, it
		// seems to be possible to pass in multiple values for the same key in
		// the web request url.  With the approach below, we will take the last
		// value for duplicate keys that are passed in on the url.
		// example:  api/playbook-dispatcher/v1/runs?filter[labels][bar]=5678&filter[labels][bar]=1234"
		for _, value := range values {
			labels[key] = value
		}
	}

	if len(labels) == 0 {
		return queryBuilder, nil
	}

	labelsJson, err := json.Marshal(labels)
	if err != nil {
		// log the error but eat it?? or throw an error all the way back out to
		// the user out??  Probably should throw it all the way back
		return queryBuilder, fmt.Errorf("unable to marshal labels into json: %w", err)
	}

	queryBuilder.Where("runs.labels @> ?", string(labelsJson))
	fmt.Println("labels json: ", string(labelsJson))

	return queryBuilder, nil
}
