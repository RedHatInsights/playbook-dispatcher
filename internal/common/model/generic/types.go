package generic

import "github.com/google/uuid"

type RunInput struct {
	Recipient     uuid.UUID
	Account       *string
	Url           string
	Hosts         []RunHostsInput
	Labels        map[string]string
	Timeout       *int
	OrgId         *string
	SatId         *uuid.UUID
	SatOrgId      *string
	Name          *string
	WebConsoleUrl *string
	Principal     *string
}

type CancelInput struct {
	Account   *string
	RunId     uuid.UUID
	OrgId     string
	Principal string
}

type RunHostsInput struct {
	AnsibleHost *string
	InventoryId *uuid.UUID
}
