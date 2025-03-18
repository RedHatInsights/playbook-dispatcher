// Code generated by github.com/atombender/go-jsonschema, DO NOT EDIT.

package message

import "encoding/json"
import "fmt"
import "time"

type PlaybookRunResponseMessageYaml struct {
	// B64Identity corresponds to the JSON schema field "b64_identity".
	B64Identity string `json:"b64_identity" yaml:"b64_identity" mapstructure:"b64_identity"`

	// Events corresponds to the JSON schema field "events".
	Events []PlaybookRunResponseMessageYamlEventsElem `json:"events" yaml:"events" mapstructure:"events"`

	// OrgId corresponds to the JSON schema field "org_id".
	OrgId string `json:"org_id" yaml:"org_id" mapstructure:"org_id"`

	// RequestId corresponds to the JSON schema field "request_id".
	RequestId string `json:"request_id" yaml:"request_id" mapstructure:"request_id"`

	// UploadTimestamp corresponds to the JSON schema field "upload_timestamp".
	UploadTimestamp time.Time `json:"upload_timestamp" yaml:"upload_timestamp" mapstructure:"upload_timestamp"`
}

type PlaybookRunResponseMessageYamlEventsElem struct {
	// Counter corresponds to the JSON schema field "counter".
	Counter int `json:"counter" yaml:"counter" mapstructure:"counter"`

	// EndLine corresponds to the JSON schema field "end_line".
	EndLine int `json:"end_line" yaml:"end_line" mapstructure:"end_line"`

	// Event corresponds to the JSON schema field "event".
	Event string `json:"event" yaml:"event" mapstructure:"event"`

	// EventData corresponds to the JSON schema field "event_data".
	EventData *PlaybookRunResponseMessageYamlEventsElemEventData `json:"event_data,omitempty" yaml:"event_data,omitempty" mapstructure:"event_data,omitempty"`

	// StartLine corresponds to the JSON schema field "start_line".
	StartLine int `json:"start_line" yaml:"start_line" mapstructure:"start_line"`

	// Stdout corresponds to the JSON schema field "stdout".
	Stdout *string `json:"stdout,omitempty" yaml:"stdout,omitempty" mapstructure:"stdout,omitempty"`

	// Uuid corresponds to the JSON schema field "uuid".
	Uuid string `json:"uuid" yaml:"uuid" mapstructure:"uuid"`
}

type PlaybookRunResponseMessageYamlEventsElemEventData struct {
	// CrcDispatcherCorrelationId corresponds to the JSON schema field
	// "crc_dispatcher_correlation_id".
	CrcDispatcherCorrelationId *string `json:"crc_dispatcher_correlation_id,omitempty" yaml:"crc_dispatcher_correlation_id,omitempty" mapstructure:"crc_dispatcher_correlation_id,omitempty"`

	// CrcDispatcherErrorCode corresponds to the JSON schema field
	// "crc_dispatcher_error_code".
	CrcDispatcherErrorCode *string `json:"crc_dispatcher_error_code,omitempty" yaml:"crc_dispatcher_error_code,omitempty" mapstructure:"crc_dispatcher_error_code,omitempty"`

	// CrcDispatcherErrorDetails corresponds to the JSON schema field
	// "crc_dispatcher_error_details".
	CrcDispatcherErrorDetails *string `json:"crc_dispatcher_error_details,omitempty" yaml:"crc_dispatcher_error_details,omitempty" mapstructure:"crc_dispatcher_error_details,omitempty"`

	// Host corresponds to the JSON schema field "host".
	Host *string `json:"host,omitempty" yaml:"host,omitempty" mapstructure:"host,omitempty"`

	// Playbook corresponds to the JSON schema field "playbook".
	Playbook *string `json:"playbook,omitempty" yaml:"playbook,omitempty" mapstructure:"playbook,omitempty"`

	// PlaybookUuid corresponds to the JSON schema field "playbook_uuid".
	PlaybookUuid *string `json:"playbook_uuid,omitempty" yaml:"playbook_uuid,omitempty" mapstructure:"playbook_uuid,omitempty"`
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *PlaybookRunResponseMessageYamlEventsElemEventData) UnmarshalJSON(b []byte) error {
	type Plain PlaybookRunResponseMessageYamlEventsElemEventData
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	if plain.Playbook != nil && len(*plain.Playbook) < 1 {
		return fmt.Errorf("field %s length: must be >= %d", "playbook", 1)
	}
	*j = PlaybookRunResponseMessageYamlEventsElemEventData(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *PlaybookRunResponseMessageYamlEventsElem) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["counter"]; raw != nil && !ok {
		return fmt.Errorf("field counter in PlaybookRunResponseMessageYamlEventsElem: required")
	}
	if _, ok := raw["end_line"]; raw != nil && !ok {
		return fmt.Errorf("field end_line in PlaybookRunResponseMessageYamlEventsElem: required")
	}
	if _, ok := raw["event"]; raw != nil && !ok {
		return fmt.Errorf("field event in PlaybookRunResponseMessageYamlEventsElem: required")
	}
	if _, ok := raw["start_line"]; raw != nil && !ok {
		return fmt.Errorf("field start_line in PlaybookRunResponseMessageYamlEventsElem: required")
	}
	if _, ok := raw["uuid"]; raw != nil && !ok {
		return fmt.Errorf("field uuid in PlaybookRunResponseMessageYamlEventsElem: required")
	}
	type Plain PlaybookRunResponseMessageYamlEventsElem
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	if 0 > plain.EndLine {
		return fmt.Errorf("field %s: must be >= %v", "end_line", 0)
	}
	if len(plain.Event) < 3 {
		return fmt.Errorf("field %s length: must be >= %d", "event", 3)
	}
	if len(plain.Event) > 50 {
		return fmt.Errorf("field %s length: must be <= %d", "event", 50)
	}
	if 0 > plain.StartLine {
		return fmt.Errorf("field %s: must be >= %v", "start_line", 0)
	}
	*j = PlaybookRunResponseMessageYamlEventsElem(plain)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *PlaybookRunResponseMessageYaml) UnmarshalJSON(b []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	if _, ok := raw["b64_identity"]; raw != nil && !ok {
		return fmt.Errorf("field b64_identity in PlaybookRunResponseMessageYaml: required")
	}
	if _, ok := raw["events"]; raw != nil && !ok {
		return fmt.Errorf("field events in PlaybookRunResponseMessageYaml: required")
	}
	if _, ok := raw["org_id"]; raw != nil && !ok {
		return fmt.Errorf("field org_id in PlaybookRunResponseMessageYaml: required")
	}
	if _, ok := raw["request_id"]; raw != nil && !ok {
		return fmt.Errorf("field request_id in PlaybookRunResponseMessageYaml: required")
	}
	if _, ok := raw["upload_timestamp"]; raw != nil && !ok {
		return fmt.Errorf("field upload_timestamp in PlaybookRunResponseMessageYaml: required")
	}
	type Plain PlaybookRunResponseMessageYaml
	var plain Plain
	if err := json.Unmarshal(b, &plain); err != nil {
		return err
	}
	*j = PlaybookRunResponseMessageYaml(plain)
	return nil
}
