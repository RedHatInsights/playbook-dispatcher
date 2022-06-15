package middleware

import (
	"net/http"
	"playbook-dispatcher/internal/common/utils"
	"strings"
)

func StoreAPIVersion(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/v1/") {
			ctx := utils.WithApiVersion(r.Context(), "v1")
			next.ServeHTTP(w, r.WithContext(ctx))
		} else if strings.Contains(r.URL.Path, "/v2/") {
			ctx := utils.WithApiVersion(r.Context(), "v2")
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			ctx := utils.WithApiVersion(r.Context(), "v1")
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	})
}
