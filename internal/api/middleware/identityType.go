package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/redhatinsights/platform-go-middlewares/v2/identity"
)

const userType = "user"
const serviceAccountType = "serviceaccount"

func EnforceIdentityType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		xrhid := identity.GetIdentity(r.Context())

		// In v2, GetIdentity returns empty XRHID when not present in context
		// Check for this case to return 500 (infrastructure issue) vs 403 (authorization issue)
		if xrhid.Identity.Type == "" {
			http.Error(w, "identity header missing in context", 500)
			return
		}

		principalType := strings.ToLower(xrhid.Identity.Type)

		if principalType != userType && principalType != serviceAccountType {
			http.Error(w, fmt.Sprintf("unauthorized principal type: %s", principalType), 403)
			return
		}

		next.ServeHTTP(w, r)
	})
}
