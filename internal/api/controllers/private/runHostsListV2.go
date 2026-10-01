// Coded in collaboration with AI
package private

import (
	"encoding/json"
	"fmt"
	"net/http"
	"playbook-dispatcher/internal/api/controllers/public"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/api/middleware"
	dbModel "playbook-dispatcher/internal/common/model/db"
	"playbook-dispatcher/internal/common/utils"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	identityMiddleware "github.com/redhatinsights/platform-go-middlewares/v2/identity"
	"gorm.io/gorm"
)

func (apii *controllers) ApiInternalV2RunHostsList(ctx echo.Context, params ApiInternalV2RunHostsListParams) error {
	identity := identityMiddleware.GetIdentity(ctx.Request().Context())

	// Blocklist check for org_id
	if utils.IsOrgIdBlocklisted(apii.config, identity.Identity.OrgID) {
		utils.GetLogFromEcho(ctx).Debugw("Rejecting request because the org_id is blocklisted")
		return ctx.NoContent(http.StatusForbidden)
	}

	limit := getLimit(params.Limit)
	offset := getOffset(params.Offset)

	fields, err := parseFields(middleware.GetDeepObject(ctx, "fields"), "data", runHostFields, defaultRunHostFields)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	queryBuilder := apii.database.
		WithContext(ctx.Request().Context()).
		Table("run_hosts").
		Joins("INNER JOIN runs on runs.id = run_hosts.run_id").
		Where("runs.org_id = ?", identity.Identity.OrgID)

	// NOTE: Removed GetAllowedServices() check - internal endpoint does not filter by service

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
			queryBuilder, err = addLabelFilterToQueryAsWhereClause(queryBuilder, labelFilters)
			if err != nil {
				instrumentation.PlaybookApiRequestError(ctx, err)
				return echo.NewHTTPError(http.StatusInternalServerError, "Unable to handle labels query!")
			}
		}

		if params.Filter.InventoryId != nil {
			inventoryId, err := uuid.Parse(*params.Filter.InventoryId)
			if err != nil {
				instrumentation.PlaybookApiRequestError(ctx, err)
				return echo.NewHTTPError(http.StatusBadRequest, "Unable to parse inventory id!")
			}

			queryBuilder.Where("run_hosts.inventory_id = ?", inventoryId)
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

	hosts := []public.RunHost{}

	for _, host := range dbRunHosts {

		runHost := public.RunHost{}
		runStatus := public.RunStatus(host.Status)

		for _, field := range fields {
			switch field {
			case fieldHost:
				runHost.Host = utils.StringRef(host.Host)
			case fieldStdout:
				runHost.Stdout = utils.StringRef(host.Log)
			case fieldStatus:
				runHost.Status = &runStatus
			case fieldRun:
				runHost.Run = &public.Run{
					Id: &host.RunID,
				}
			case fieldLinks:
				runHost.Links = &public.RunHostLinks{
					InventoryHost: inventoryLink(host.InventoryID),
				}
			case fieldInventoryId:
				if host.InventoryID != nil {
					runHost.InventoryId = host.InventoryID
				}
			}
		}

		hosts = append(hosts, runHost)
	}

	return ctx.JSON(http.StatusOK, &public.RunHosts{
		Data: hosts,
		Meta: public.Meta{
			Count: len(hosts),
			Total: int(total),
		},
		Links: createLinks("/internal/v2/run_hosts", middleware.GetQueryString(ctx), getLimit(params.Limit), getOffset(params.Offset), int(total)),
	})
}

// Helper functions duplicated from public package

const (
	fieldHost        = "host"
	fieldRun         = "run"
	fieldStatus      = "status"
	fieldStdout      = "stdout"
	fieldLinks       = "links"
	fieldInventoryId = "inventory_id"
)

var (
	runHostFields        = utils.IndexStrings(fieldHost, fieldRun, fieldStatus, fieldStdout, fieldLinks, fieldInventoryId)
	defaultRunHostFields = []string{fieldHost, fieldRun, fieldStatus}
)

const defaultLimit = 50

func getLimit(limit *public.Limit) int {
	if limit != nil {
		return (int(*limit))
	}
	return defaultLimit
}

func getOffset(offset *public.Offset) int {
	if offset != nil {
		return int(*offset)
	}
	return 0
}

func parseFields(input map[string][]string, key string, knownFields map[string]string, defaults []string) ([]string, error) {
	selectedFields, ok := input[key]

	if !ok {
		return defaults, nil
	}

	result := []string{}

	for _, value := range selectedFields {
		for _, field := range strings.Split(value, ",") {
			if _, ok := knownFields[field]; ok {
				result = append(result, field)
			} else {
				return nil, fmt.Errorf("unknown field: %s", field)
			}
		}
	}

	return result, nil
}

func createLinks(base string, queryString string, limit, offset, total int) public.Links {
	lastPage := int(utils.Max(total-1, 0) / limit)

	links := public.Links{
		First: createLink(base, queryString, limit, 0),
		Last:  createLink(base, queryString, limit, lastPage*limit),
	}

	if offset > 0 {
		previous := createLink(base, queryString, limit, utils.Max(offset-limit, 0))
		links.Previous = &previous
	}

	if offset+limit < total {
		next := createLink(base, queryString, limit, offset+limit)
		links.Next = &next
	}

	return links
}

func createLink(base string, queryString string, limit, offset int) string {
	if queryString == "" {
		return fmt.Sprintf("%s?limit=%d&offset=%d", base, limit, offset)
	}
	return fmt.Sprintf("%s?%s&limit=%d&offset=%d", base, queryString, limit, offset)
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

func addLabelFilterToQueryAsWhereClause(queryBuilder *gorm.DB, labelFilters map[string][]string) (*gorm.DB, error) {
	labels := make(map[string]string)

	for key, values := range labelFilters {
		for _, value := range values {
			labels[key] = value
		}
	}

	if len(labels) == 0 {
		return queryBuilder, nil
	}

	labelsJson, err := json.Marshal(labels)
	if err != nil {
		return queryBuilder, fmt.Errorf("unable to marshal labels into json: %w", err)
	}

	queryBuilder.Where("runs.labels @> ?", string(labelsJson))

	return queryBuilder, nil
}
