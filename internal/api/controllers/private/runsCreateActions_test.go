package private

import (
	"errors"
	"net/http"
	"testing"

	"playbook-dispatcher/internal/api/controllers/public"
	"playbook-dispatcher/internal/api/dispatch"
	"playbook-dispatcher/internal/common/model/generic"
	"playbook-dispatcher/internal/common/utils"

	"github.com/RedHatInsights/tenant-utils/pkg/tenantid"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func TestHandleRunCreateError(t *testing.T) {
	tests := []struct {
		name         string
		err          error
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "RecipientNotFoundError returns 404",
			err:          &dispatch.RecipientNotFoundError{},
			expectedCode: http.StatusNotFound,
			expectedMsg:  "Receipient not found",
		},
		{
			name:         "TenantNotFoundError returns 404",
			err:          &tenantid.TenantNotFoundError{},
			expectedCode: http.StatusNotFound,
			expectedMsg:  "Tenant not found",
		},
		{
			name:         "BlocklistedOrgIdError returns 400",
			err:          &utils.BlocklistedOrgIdError{},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "Block listed org",
		},
		{
			name:         "Unknown error returns 500",
			err:          errors.New("some other error"),
			expectedCode: http.StatusInternalServerError,
			expectedMsg:  "Unexpected error during processing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleRunCreateError(tt.err)
			if result.Code != tt.expectedCode {
				t.Errorf("handleRunCreateError(%T) = %d, want %d", tt.err, result.Code, tt.expectedCode)
			}
			if result.Message == nil || *result.Message != tt.expectedMsg {
				t.Errorf("handleRunCreateError(%T) message = %v, want %v", tt.err, result.Message, tt.expectedMsg)
			}
		})
	}
}

func TestRunInputV2GenericMap(t *testing.T) {
	recipient := uuid.New()
	orgId := "12345"
	url := "http://example.com"
	name := "playbook"
	webConsoleUrl := "http://console.example.com"
	principal := "test_user"
	timeout := 3600
	inventoryId := uuid.New()
	satId := uuid.New()
	satIdString := satId.String()
	satOrgId := "sat-org-1"

	hosts := []generic.RunHostsInput{
		{InventoryId: &inventoryId},
	}

	runInput := RunInputV2{
		Recipient:     public.RunRecipient(recipient),
		OrgId:         public.OrgId(orgId),
		Url:           public.Url(url),
		Name:          public.PlaybookName(name),
		WebConsoleUrl: (*public.WebConsoleUrl)(&webConsoleUrl),
		Principal:     Principal(principal),
		Timeout:       (*public.RunTimeout)(&timeout),
		Hosts:         nil, // not used in mapping, we pass parsedHosts
		RecipientConfig: &RecipientConfig{
			SatId:    &satIdString,
			SatOrgId: &satOrgId,
		},
		Labels: &public.Labels{"foo": "bar"},
	}

	cfg := viper.New()
	result := RunInputV2GenericMap(runInput, recipient, hosts, &satId, cfg)

	if result.Recipient != recipient {
		t.Errorf("Recipient: got %v, want %v", result.Recipient, recipient)
	}
	if result.OrgId != orgId {
		t.Errorf("OrgId: got %v, want %v", result.OrgId, orgId)
	}
	if result.Url != url {
		t.Errorf("Url: got %v, want %v", result.Url, url)
	}
	if result.Name == nil || *result.Name != name {
		t.Errorf("Name: got %v, want %v", result.Name, name)
	}
	if result.WebConsoleUrl == nil || *result.WebConsoleUrl != webConsoleUrl {
		t.Errorf("WebConsoleUrl: got %v, want %v", result.WebConsoleUrl, webConsoleUrl)
	}
	if result.Principal == nil || *result.Principal != principal {
		t.Errorf("Principal: got %v, want %v", result.Principal, principal)
	}
	if result.Timeout == nil || *result.Timeout != timeout {
		t.Errorf("Timeout: got %v, want %v", result.Timeout, timeout)
	}
	if len(result.Hosts) != 1 || result.Hosts[0].InventoryId == nil || *result.Hosts[0].InventoryId != inventoryId {
		t.Errorf("Hosts: got %v, want inventoryId %v", result.Hosts, inventoryId)
	}
	if result.SatId == nil || *result.SatId != satId {
		t.Errorf("SatId: got %v, want %v", result.SatId, satId)
	}
	if result.SatOrgId == nil || *result.SatOrgId != satOrgId {
		t.Errorf("SatOrgId: got %v, want %v", result.SatOrgId, satOrgId)
	}
	if v, ok := result.Labels["foo"]; !ok || v != "bar" {
		t.Errorf("Labels: got %v, want foo=bar", result.Labels)
	}
}
