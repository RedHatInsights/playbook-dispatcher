package message

import "time"

type IngressValidationRequest struct {
	Account     string            `json:"account"`
	Category    string            `json:"category"`
	Metadata    map[string]string `json:"metadata"`
	RequestID   string            `json:"request_id"`
	Principal   string            `json:"principal"`
	Service     string            `json:"service"`
	Size        int64             `json:"size"`
	URL         string            `json:"url"`
	ID          string            `json:"id,omitempty"`
	B64Identity string            `json:"b64_identity"`
	Timestamp   time.Time         `json:"timestamp"`
}

type IngressValidationDetails struct {
	Validation string `json:"validation"`
	Reason string `json:"reason"`
	Reporter string `json:"reporter"`
}

type IngressValidationResponse struct {
	IngressValidationRequest
	IngressValidationDetails
	SystemID string `json:"system_id"`
}
