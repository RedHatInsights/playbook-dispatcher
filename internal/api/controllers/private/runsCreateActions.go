package private

import (
	"fmt"
	"net/http"

	"playbook-dispatcher/internal/api/controllers/public"
	"playbook-dispatcher/internal/api/dispatch"
	"playbook-dispatcher/internal/common/model/generic"
	"playbook-dispatcher/internal/common/utils"

	"github.com/RedHatInsights/tenant-utils/pkg/tenantid"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"
)

func getLabels(input *public.Labels) map[string]string {
	if input == nil {
		return map[string]string{}
	}

	return input.AdditionalProperties
}

// this will panic if the given value is not a valid UUID
// this function should only be used on values that have been validated as UUID using the openapi middleware
func parseValidatedUUID(value string) uuid.UUID {
	return uuid.MustParse(value)
}

func parseRunHosts(input *RunInputHosts) []generic.RunHostsInput {
	if input == nil {
		return []generic.RunHostsInput{}
	}

	result := make([]generic.RunHostsInput, len(*input))

	for i, host := range *input {
		result[i] = generic.RunHostsInput{
			AnsibleHost: host.AnsibleHost,
		}

		if host.InventoryId != nil {
			result[i].InventoryId = utils.UUIDRef(parseValidatedUUID(*host.InventoryId))
		}

	}

	return result
}

func RunInputV1GenericMap(runInput RunInput, orgId *string, parsedRecipient uuid.UUID, parsedHosts []generic.RunHostsInput, cfg *viper.Viper) generic.RunInput {
	return generic.RunInput{
		Recipient: parsedRecipient,
		OrgId:     orgId,
		Account:   string(runInput.Account),
		Url:       string(runInput.Url),
		Labels:    getLabels(runInput.Labels),
		Timeout:   (*int)(runInput.Timeout),
		Hosts:     parsedHosts,
	}
}

func RunInputV2GenericMap(
	runInput RunInputV2,
	account string,
	parsedRecipient uuid.UUID,
	parsedHosts []generic.RunHostsInput,
	parsedSatID *uuid.UUID,
	cfg *viper.Viper,
) generic.RunInput {
	orgIdString := string(runInput.OrgId)
	playbookName := string(runInput.Name)
	principal := string(runInput.Principal)

	result := generic.RunInput{
		Recipient:     parsedRecipient,
		OrgId:         &orgIdString,
		Account:       account,
		Url:           string(runInput.Url),
		Labels:        getLabels(runInput.Labels),
		Timeout:       (*int)(runInput.Timeout),
		Hosts:         parsedHosts,
		Name:          &playbookName,
		WebConsoleUrl: (*string)(runInput.WebConsoleUrl),
		Principal:     &principal,
		SatId:         parsedSatID,
	}

	if runInput.RecipientConfig != nil {
		result.SatOrgId = runInput.RecipientConfig.SatOrgId
	}

	return result
}

func validateSatelliteFields(runInput RunInputV2) error {
	if runInput.RecipientConfig == nil {
		return nil
	}

	if (runInput.RecipientConfig.SatId == nil) != (runInput.RecipientConfig.SatOrgId == nil) {
		return fmt.Errorf("Both sat_id and sat_org need to be defined")
	}

	if runInput.RecipientConfig.SatId == nil {
		return nil
	}

	if runInput.Hosts == nil {
		return fmt.Errorf("Hosts need to be defined")
	}

	if len(*runInput.Hosts) == 0 {
		return fmt.Errorf("Hosts cannot be empty")
	}

	for _, host := range *runInput.Hosts {
		if host.InventoryId == nil {
			return fmt.Errorf("Inventory ID needs to be defined")
		}
	}

	return nil
}

func runCreateError(code int) *RunCreated {
	return &RunCreated{
		Code: code,
		// TODO report error
	}
}

func handleRunCreateError(err error) *RunCreated {
	if _, ok := err.(*dispatch.RecipientNotFoundError); ok {
		return runCreateError(http.StatusNotFound)
	}

	if _, ok := err.(*tenantid.TenantNotFoundError); ok {
		return runCreateError(http.StatusNotFound)
	}

	return runCreateError(http.StatusInternalServerError)
}

func runCreated(runID uuid.UUID) *RunCreated {
	return &RunCreated{
		Code: http.StatusCreated,
		Id:   (*public.RunId)(utils.StringRef(runID.String())),
	}
}

func invalidRequest(ctx echo.Context, err error) error {
	return ctx.JSON(http.StatusBadRequest, Error{
		Message: err.Error(),
	})
}
