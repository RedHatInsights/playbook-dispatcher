package message

import "time"

type IngressValidationRequest struct {
	Account     string            `json:"account"`
	Category    string            `json:"category"`
	Metadata    map[string]string `json:"metadata"`
	RequestID   string            `json:"request_id"`
	OrgID       string            `json:"org_id"`
	Service     string            `json:"service"`
	Size        int64             `json:"size"`
	URL         string            `json:"url"`
	ID          string            `json:"id,omitempty"`
	B64Identity string            `json:"b64_identity"`
	Timestamp   time.Time         `json:"timestamp"`
}

type IngressValidationResponse struct {
	IngressValidationRequest
	Validation string `json:"validation"`
}

func NewResponse(req *IngressValidationRequest, result string) *IngressValidationResponse {
	return &IngressValidationResponse{
		IngressValidationRequest: *req,
		Validation:               result,
	}
}
