package private

import (
	"net/http"
	"playbook-dispatcher/internal/api/instrumentation"
	"playbook-dispatcher/internal/api/middleware"
	"playbook-dispatcher/internal/common/model/generic"
	"playbook-dispatcher/internal/common/utils"

	"github.com/RedHatInsights/tenant-utils/pkg/tenantid"

	"github.com/labstack/echo/v4"
)

//go:generate fungen -types RunInputV3,RunsCreatedV3 -methods PMap -package private -filename utils.v3.gen.go
func (this *controllers) ApiInternalV3RunsCreate(ctx echo.Context) error {
	var input RunInputV3List

	err := utils.ReadRequestBody(ctx, &input)
	if err != nil {
		utils.GetLogFromEcho(ctx).Error(err)
		return ctx.NoContent(http.StatusBadRequest)
	}

	// process individual requests concurrently
	result := input.PMapRunsCreatedV3(func(runInputV3 RunInputV3) RunsCreatedV3 {
		runsCreated := make(RunsCreatedV3, len(runInputV3.Hosts))
		context := utils.WithOrgId(ctx.Request().Context(), string(runInputV3.OrgId))
		context = utils.WithRequestType(context, getRequestTypeLabelV3(runInputV3))

		// translate org_id to EAN
		ean, err := this.translator.OrgIDToEAN(ctx.Request().Context(), string(runInputV3.OrgId))
		if err != nil {
			if _, ok := err.(*tenantid.TenantNotFoundError); ok {
				return runCreateError(http.StatusNotFound)
			}

			utils.GetLogFromEcho(ctx).Error(err)
			return runCreateError(http.StatusInternalServerError)
		}

		if ean == nil {
			utils.GetLogFromEcho(ctx).Warnw("Anemic tenant not supported", "org_id", string(runInputV3.OrgId))
			return runCreateError(http.StatusBadRequest)
		}

		hosts := parseRunHosts(&runInputV3.Hosts)  // database model generic.RunHostsInput

		// request the Inventory details for each of these hosts.
		recipientForHost := resolveRecipientIds(hosts)
		// now dispatch a run for each of the hosts in this batch
		for i, host := range(hosts) {
			recipientUUID := parseValidatedUUID(recipientForHost[host])
			var thisHost = []generic.RunHostsInput{host}  // single host in list
			runInput := RunInputV3GenericMap(runInputV3, recipientUUID, thisHost, this.config)
			runID, _, err := this.dispatchManager.ProcessRun(context, *ean, middleware.GetPSKPrincipal(context), runInput, "v3")
			runsCreated[i] = RunCreatedV3{
				Code: err.code,
				InventoryId: host.InventoryId,
				RunId: runID,
			}
		}

		return runsCreated
	})

	return ctx.JSON(http.StatusMultiStatus, result)
}

// the structure returned from Inventory
type InventoryHostResponse struct {
	// the fields we have to parse and that we care about:
	// /hosts/{host_id_list} :
	// {
	//  "results": [
	//    {
	//      "id": "db297f73-3d17-47a9-829e-86f4e3da4205",
	//      "facts": [
	//        {
	//          "namespace": "satellite",
	//          "facts": {
	//        *** "satellite_instance_id": "3e9f5a3b-d2fc-47a7-af49-eb8645f58551", ***
	//          }
	//        }
	//      ]
	//    }
	//  ]
	//}
	Results [] struct {
		Id string `json:"id"`
		Facts [] struct {
			Namespace string `json:"namespace"`
			NamespaceFacts [] struct {
				SatelliteInstanceId string `json:"satellite_instance_id"`
			} `json:"facts"`
		} `json:"facts"`
	} `json:"results"`
}

type InventorySystemProfileResponse struct{
	// /hosts/{host_id_list}/system_profile?fields[system_profile]=rhc_client_id:
	//{
	//  "results": [
	//    {
	//      "id": "aa3b4d08-f3f0-40af-949f-2d5168b4d090",
	//      "system_profile": {
	//        "rhc_client_id": "6c741de1-c67a-4c70-a9e9-093781b2ee26",
	//??        "rhc_config_state": "481c0739-0cfc-4e73-bc62-60d240343839",
	//      }
	//    },
	Results [] struct {
		Id string `json:"id"`
		SystemProfile struct {
			RhcClientId string `json:"rhc_client_id"`
			// RhcConfigState string `json:"rhc_config_state"`
		} `json:"system_profile"`
	} `json:"results"`
}


func resolveRecipientIds([]generic.RunHostsInput) map[generic.RunHostsInput]string {
	// This creates a mapping from Inventory host UUID to Recipient ID:
	//   * if the host is self-managed, the recipient is the `rhc_client_id`
	//     in the `system_profile`
	//   * if the host is managed by a Satellite, the recipient is the
	//     `satellite_instance_id` in the `satellite` fact namespace.
	// a. it'd be good if we could just use the syndicated Inventory
	//    table here, but we don't have that available.
	// b. so instead we're going to have to use the Inventory API, which
	//    is a problem for a helper library.  We get back a mapping from
	//    the Inventory UUID to the recipient UUID.
	recipientForHost := make(map[RunHostsInput]string)
	// temporary stuff for now
	for _, host := range(hosts) {
		recipientForHost[host] = "Foo"
	}
	// real algorithm would be something like:
	// 1. split hosts up into batches of ~50 or 100, due to host UUIDs
	//    being specified on the URL path
	// 2. request each batch from both hosts and system_profile endpoints
	//   a. hosts that have rhc_client_id defined in their system_profile
	//      get that set as their recipient ID
	//   b. hosts that have the satellite_instance_id set in the 'satellite'
	//      namespace get that set as their recipient ID
	// 3. If the host doesn't occur in either, we can't dispatch to it, so
	//    it doesn't appear in the map
	return recipientForHost
}

func runCreateError(code int) *RunCreated {
	return &RunCreatedV3{
		Code: code,
		// TODO report error
	}
}

func getRequestTypeLabelV3(run RunInputV3) string {
	result := instrumentation.LabelAnsibleRequest

	if run.RecipientConfig != nil && run.RecipientConfig.SatId != nil {
		result = instrumentation.LabelSatRequest
	}

	return result
}
