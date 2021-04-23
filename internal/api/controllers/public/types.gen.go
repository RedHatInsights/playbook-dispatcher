// Package public provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package public

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// Account defines model for Account.
type Account string

// CreatedAt defines model for CreatedAt.
type CreatedAt time.Time

// Error defines model for Error.
type Error struct {
	Message string `json:"message"`
}

// Labels defines model for Labels.
type Labels struct {
	AdditionalProperties map[string]string `json:"-"`
}

// Links defines model for Links.
type Links struct {

	// relative link to the first page of the query results
	First string `json:"first"`

	// relative link to the last page of the query results
	Last string `json:"last"`

	// relative link to the next page of the query results
	Next *string `json:"next,omitempty"`

	// relative link to the previous page of the query results
	Previous *string `json:"previous,omitempty"`
}

// Meta defines model for Meta.
type Meta struct {

	// number of results returned
	Count int `json:"count"`

	// total number of results matching the query
	Total int `json:"total"`
}

// Run defines model for Run.
type Run struct {

	// Identifier of the tenant
	Account *Account `json:"account,omitempty"`

	// Unique identifier used to match work request with responses
	CorrelationId *RunCorrelationId `json:"correlation_id,omitempty"`

	// A timestamp when the entry was created
	CreatedAt *CreatedAt `json:"created_at,omitempty"`

	// Unique identifier of a Playbook run
	Id *RunId `json:"id,omitempty"`

	// Additional metadata about the Playbook run. Can be used for filtering purposes.
	Labels *Labels `json:"labels,omitempty"`

	// Identifier of the host to which a given Playbook is addressed
	Recipient *RunRecipient `json:"recipient,omitempty"`

	// Service that triggered the given Playbook run
	Service *Service `json:"service,omitempty"`

	// Current status of a Playbook run
	Status *RunStatus `json:"status,omitempty"`

	// Amount of seconds after which the run is considered failed due to timeout
	Timeout *RunTimeout `json:"timeout,omitempty"`

	// A timestamp when the entry was last updated
	UpdatedAt *UpdatedAt `json:"updated_at,omitempty"`

	// URL hosting the Playbook
	Url *Url `json:"url,omitempty"`
}

// RunCorrelationId defines model for RunCorrelationId.
type RunCorrelationId string

// RunHost defines model for RunHost.
type RunHost struct {

	// Name used to identify a host within Ansible inventory
	Host  *string       `json:"host,omitempty"`
	Links *RunHostLinks `json:"links,omitempty"`
	Run   *Run          `json:"run,omitempty"`

	// Current status of a Playbook run
	Status *RunStatus `json:"status,omitempty"`

	// Output produced by running Ansible Playbook on the given host
	Stdout *string `json:"stdout,omitempty"`
}

// RunHostLinks defines model for RunHostLinks.
type RunHostLinks struct {
	InventoryHost *string `json:"inventory_host"`
}

// RunHosts defines model for RunHosts.
type RunHosts struct {
	Data  []RunHost `json:"data"`
	Links Links     `json:"links"`

	// Information about returned entities
	Meta Meta `json:"meta"`
}

// RunId defines model for RunId.
type RunId string

// RunLabelsNullable defines model for RunLabelsNullable.
type RunLabelsNullable struct {
	AdditionalProperties map[string]string `json:"-"`
}

// RunRecipient defines model for RunRecipient.
type RunRecipient string

// RunStatus defines model for RunStatus.
type RunStatus string

// List of RunStatus
const (
	RunStatus_failure RunStatus = "failure"
	RunStatus_running RunStatus = "running"
	RunStatus_success RunStatus = "success"
	RunStatus_timeout RunStatus = "timeout"
)

// RunTimeout defines model for RunTimeout.
type RunTimeout int

// Runs defines model for Runs.
type Runs struct {
	Data  []Run `json:"data"`
	Links Links `json:"links"`

	// Information about returned entities
	Meta Meta `json:"meta"`
}

// Service defines model for Service.
type Service string

// ServiceNullable defines model for ServiceNullable.
type ServiceNullable string

// StatusNullable defines model for StatusNullable.
type StatusNullable string

// List of StatusNullable
const (
	StatusNullable_failure StatusNullable = "failure"
	StatusNullable_running StatusNullable = "running"
	StatusNullable_success StatusNullable = "success"
	StatusNullable_timeout StatusNullable = "timeout"
)

// UpdatedAt defines model for UpdatedAt.
type UpdatedAt time.Time

// Url defines model for Url.
type Url string

// Limit defines model for Limit.
type Limit int

// Offset defines model for Offset.
type Offset int

// RunHostFields defines model for RunHostFields.
type RunHostFields struct {
	Data *[]string `json:"data,omitempty"`
}

// RunHostFilter defines model for RunHostFilter.
type RunHostFilter struct {
	InventoryId *string `json:"inventory_id"`
	Run         *struct {
		Id      *string            `json:"id"`
		Labels  *RunLabelsNullable `json:"labels"`
		Service *ServiceNullable   `json:"service"`
	} `json:"run"`
	Status *StatusNullable `json:"status"`
}

// RunsFields defines model for RunsFields.
type RunsFields struct {
	Data *[]string `json:"data,omitempty"`
}

// RunsFilter defines model for RunsFilter.
type RunsFilter struct {
	Labels    *RunLabelsNullable `json:"labels"`
	Recipient *string            `json:"recipient"`
	Service   *ServiceNullable   `json:"service"`
	Status    *StatusNullable    `json:"status"`
}

// RunsSortBy defines model for RunsSortBy.
type RunsSortBy string

// List of RunsSortBy
const (
	RunsSortBy_created_at      RunsSortBy = "created_at"
	RunsSortBy_created_at_asc  RunsSortBy = "created_at:asc"
	RunsSortBy_created_at_desc RunsSortBy = "created_at:desc"
)

// BadRequest defines model for BadRequest.
type BadRequest Error

// Forbidden defines model for Forbidden.
type Forbidden Error

// ApiRunHostsListParams defines parameters for ApiRunHostsList.
type ApiRunHostsListParams struct {

	// Allows for filtering based on various criteria
	Filter *RunHostFilter `json:"filter,omitempty"`

	// Defines fields to be returned in the response.
	Fields *RunHostFields `json:"fields,omitempty"`

	// Maximum number of results to return
	Limit *Limit `json:"limit,omitempty"`

	// Indicates the starting position of the query relative to the complete set of items that match the query
	Offset *Offset `json:"offset,omitempty"`
}

// ApiRunsListParams defines parameters for ApiRunsList.
type ApiRunsListParams struct {

	// Allows for filtering based on various criteria
	Filter *RunsFilter `json:"filter,omitempty"`

	// Defines fields to be returned in the response.
	Fields *RunsFields `json:"fields,omitempty"`

	// Sort order
	SortBy *RunsSortBy `json:"sort_by,omitempty"`

	// Maximum number of results to return
	Limit *Limit `json:"limit,omitempty"`

	// Indicates the starting position of the query relative to the complete set of items that match the query
	Offset *Offset `json:"offset,omitempty"`
}

// Getter for additional properties for Labels. Returns the specified
// element and whether it was found
func (a Labels) Get(fieldName string) (value string, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for Labels
func (a *Labels) Set(fieldName string, value string) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]string)
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for Labels to handle AdditionalProperties
func (a *Labels) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]string)
		for fieldName, fieldBuf := range object {
			var fieldVal string
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("error unmarshaling field %s", fieldName))
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for Labels to handle AdditionalProperties
func (a Labels) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error marshaling '%s'", fieldName))
		}
	}
	return json.Marshal(object)
}

// Getter for additional properties for RunLabelsNullable. Returns the specified
// element and whether it was found
func (a RunLabelsNullable) Get(fieldName string) (value string, found bool) {
	if a.AdditionalProperties != nil {
		value, found = a.AdditionalProperties[fieldName]
	}
	return
}

// Setter for additional properties for RunLabelsNullable
func (a *RunLabelsNullable) Set(fieldName string, value string) {
	if a.AdditionalProperties == nil {
		a.AdditionalProperties = make(map[string]string)
	}
	a.AdditionalProperties[fieldName] = value
}

// Override default JSON handling for RunLabelsNullable to handle AdditionalProperties
func (a *RunLabelsNullable) UnmarshalJSON(b []byte) error {
	object := make(map[string]json.RawMessage)
	err := json.Unmarshal(b, &object)
	if err != nil {
		return err
	}

	if len(object) != 0 {
		a.AdditionalProperties = make(map[string]string)
		for fieldName, fieldBuf := range object {
			var fieldVal string
			err := json.Unmarshal(fieldBuf, &fieldVal)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("error unmarshaling field %s", fieldName))
			}
			a.AdditionalProperties[fieldName] = fieldVal
		}
	}
	return nil
}

// Override default JSON handling for RunLabelsNullable to handle AdditionalProperties
func (a RunLabelsNullable) MarshalJSON() ([]byte, error) {
	var err error
	object := make(map[string]json.RawMessage)

	for fieldName, field := range a.AdditionalProperties {
		object[fieldName], err = json.Marshal(field)
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("error marshaling '%s'", fieldName))
		}
	}
	return json.Marshal(object)
}
