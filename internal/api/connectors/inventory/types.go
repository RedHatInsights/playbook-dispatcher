package inventory

type HostDetailsResponse struct {
	ID             string                 `json:"id"`
	DisplayName    string                 `json:"display_name,omitempty"`
	Facts          map[string]interface{} `json:"facts,omitempty"`
	CanonicalFacts map[string]interface{} `json:"canonical_facts,omitempty"`
}

type SystemProfileDetailsResponse struct {
	ID                 string                 `json:"id"`
	SystemProfileFacts map[string]interface{} `json:"system_profile_facts,omitempty"`
}
