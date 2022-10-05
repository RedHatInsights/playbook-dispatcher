package inventory

type HostDetails struct {
	ID             string                 `json:"id"`
	DisplayName    string                 `json:"display_name,omitempty"`
	Facts          map[string]interface{} `json:"facts,omitempty"`
	CanonicalFacts map[string]interface{} `json:"canonical_facts,omitempty"`
}

type SystemProfileDetails struct {
	ID                 string                 `json:"id"`
	SystemProfileFacts map[string]interface{} `json:"system_profile,omitempty"`
}
