package middleware

import (
	"playbook-dispatcher/internal/api/rbac"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestGetRbacAllowedServices_Empty(t *testing.T) {
	permissions := []rbac.Access{}
	result := getRbacAllowedServices(permissions)
	assert.Empty(t, result)
}

func TestLogComparison_Match(t *testing.T) {
	log := zap.NewNop().Sugar()
	rbacServices := []string{"config_manager", "remediations"}
	kesselServices := []string{"remediations", "config_manager"} // Different order

	// Should not panic
	logComparison(rbacServices, kesselServices, log)
}

func TestLogComparison_Mismatch(t *testing.T) {
	log := zap.NewNop().Sugar()
	rbacServices := []string{"config_manager", "remediations"}
	kesselServices := []string{"config_manager", "tasks"}

	// Should not panic and should log mismatch
	logComparison(rbacServices, kesselServices, log)
}

func TestLogComparison_EmptyLists(t *testing.T) {
	log := zap.NewNop().Sugar()
	rbacServices := []string{}
	kesselServices := []string{}

	// Should handle empty lists
	logComparison(rbacServices, kesselServices, log)
}

func TestLogComparison_OneEmpty(t *testing.T) {
	log := zap.NewNop().Sugar()
	rbacServices := []string{"config_manager"}
	kesselServices := []string{}

	// Should handle mismatch when one is empty
	logComparison(rbacServices, kesselServices, log)
}
