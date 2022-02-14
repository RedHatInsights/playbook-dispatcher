// Package private provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package private

import (
	externalRef0 "playbook-dispatcher/internal/api/controllers/public"
)

// Error defines model for Error.
type Error struct {

	// Human readable error message
	Message string `json:"message"`
}

// OrgId defines model for OrgId.
type OrgId string

// PlaybookName defines model for PlaybookName.
type PlaybookName string

// Principal defines model for Principal.
type Principal string

// RecipientConfig defines model for RecipientConfig.
type RecipientConfig struct {

	// Identifier of the Satellite instance
	SatId *string `json:"sat_id,omitempty"`

	// Identifier of the organization within Satellite
	SatOrgId *string `json:"sat_org_id,omitempty"`
}

// RecipientStatus defines model for RecipientStatus.
type RecipientStatus struct {
	// Embedded struct due to allOf(#/components/schemas/RecipientWithOrg)
	RecipientWithOrg `yaml:",inline"`
	// Embedded fields due to inline allOf schema

	// Indicates whether a connection is established with the recipient
	Connected bool `json:"connected"`
}

// RecipientWithOrg defines model for RecipientWithOrg.
type RecipientWithOrg struct {

	// Identifies the organization that the given resource belongs to
	OrgId OrgId `json:"org_id"`

	// Identifier of the host to which a given Playbook is addressed
	Recipient externalRef0.RunRecipient `json:"recipient"`
}

// RunCreated defines model for RunCreated.
type RunCreated struct {

	// status code of the request
	Code int `json:"code"`

	// Unique identifier of a Playbook run
	Id *externalRef0.RunId `json:"id,omitempty"`
}

// RunInput defines model for RunInput.
type RunInput struct {

	// Identifier of the tenant
	Account externalRef0.Account `json:"account"`

	// Optionally, information about hosts involved in the Playbook run can be provided.
	// This information is used to pre-allocate run_host resources.
	// Moreover, it can be used to create a connection between a run_host resource and host inventory.
	Hosts *RunInputHosts `json:"hosts,omitempty"`

	// Additional metadata about the Playbook run. Can be used for filtering purposes.
	Labels *externalRef0.Labels `json:"labels,omitempty"`

	// Identifier of the host to which a given Playbook is addressed
	Recipient externalRef0.RunRecipient `json:"recipient"`

	// Amount of seconds after which the run is considered failed due to timeout
	Timeout *externalRef0.RunTimeout `json:"timeout,omitempty"`

	// URL hosting the Playbook
	Url externalRef0.Url `json:"url"`
}

// RunInputHosts defines model for RunInputHosts.
type RunInputHosts []struct {

	// Host name as known to Ansible inventory.
	// Used to identify the host in status reports.
	AnsibleHost *string `json:"ansible_host,omitempty"`

	// Inventory id of the given host
	InventoryId *string `json:"inventory_id,omitempty"`
}

// RunInputV2 defines model for RunInputV2.
type RunInputV2 struct {

	// Optionally, information about hosts involved in the Playbook run can be provided.
	// This information is used to pre-allocate run_host resources.
	// Moreover, it can be used to create a connection between a run_host resource and host inventory.
	Hosts *RunInputHosts `json:"hosts,omitempty"`

	// Additional metadata about the Playbook run. Can be used for filtering purposes.
	Labels *externalRef0.Labels `json:"labels,omitempty"`

	// Human readable name of the playbook run. Used to present the given playbook run in external systems (Satellite).
	Name PlaybookName `json:"name"`

	// Identifier of the tenant
	OrgId externalRef0.OrgId `json:"org_id"`

	// Username of the user interacting with the service
	Principal Principal `json:"principal"`

	// Identifier of the host to which a given Playbook is addressed
	Recipient externalRef0.RunRecipient `json:"recipient"`

	// recipient-specific configuration options
	RecipientConfig *RecipientConfig `json:"recipient_config,omitempty"`

	// Amount of seconds after which the run is considered failed due to timeout
	Timeout *externalRef0.RunTimeout `json:"timeout,omitempty"`

	// URL hosting the Playbook
	Url externalRef0.Url `json:"url"`

	// URL that points to the section of the web console where the user find more information about the playbook run. The field is optional but highly suggested.
	WebConsoleUrl *WebConsoleUrl `json:"web_console_url,omitempty"`
}

// RunsCreated defines model for RunsCreated.
type RunsCreated []RunCreated

// WebConsoleUrl defines model for WebConsoleUrl.
type WebConsoleUrl string

// BadRequest defines model for BadRequest.
type BadRequest Error

// ApiInternalRunsCreateJSONBody defines parameters for ApiInternalRunsCreate.
type ApiInternalRunsCreateJSONBody []RunInput

// ApiInternalV2RunsCreateJSONBody defines parameters for ApiInternalV2RunsCreate.
type ApiInternalV2RunsCreateJSONBody []RunInputV2

// ApiInternalV2RecipientsStatusJSONBody defines parameters for ApiInternalV2RecipientsStatus.
type ApiInternalV2RecipientsStatusJSONBody []RecipientWithOrg

// ApiInternalRunsCreateRequestBody defines body for ApiInternalRunsCreate for application/json ContentType.
type ApiInternalRunsCreateJSONRequestBody ApiInternalRunsCreateJSONBody

// ApiInternalV2RunsCreateRequestBody defines body for ApiInternalV2RunsCreate for application/json ContentType.
type ApiInternalV2RunsCreateJSONRequestBody ApiInternalV2RunsCreateJSONBody

// ApiInternalV2RecipientsStatusRequestBody defines body for ApiInternalV2RecipientsStatus for application/json ContentType.
type ApiInternalV2RecipientsStatusJSONRequestBody ApiInternalV2RecipientsStatusJSONBody
