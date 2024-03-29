package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/redhatinsights/platform-go-middlewares/identity"
)

const userType = "user"
const serviceAccountType = "serviceaccount"

func EnforceIdentityType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		value := r.Context().Value(identity.Key)
		identity, ok := value.(identity.XRHID)

		if !ok {
			http.Error(w, "identity header missing in context", 500)
			return
		}

		principalType := strings.ToLower(identity.Identity.Type)

		if principalType != userType && principalType != serviceAccountType {
			http.Error(w, fmt.Sprintf("unauthorized principal type: %s", principalType), 403)
			return
		}

		next.ServeHTTP(w, r)
	})
}
