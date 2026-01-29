package middleware

import (
	"context"
	"net/http"
	"os"
	"playbook-dispatcher/internal/common/utils"
	"regexp"
	"strings"

	"github.com/labstack/echo/v5"
	"go.uber.org/zap"
)

type principalKey int

const pskPrincipal principalKey = iota

var headerMatcher = regexp.MustCompile(`^PSK\s+([0-9a-zA-Z]+)$`)
var envMatcher = regexp.MustCompile(`^PSK_AUTH_(.+?)=(.+?)$`)

func CheckPskAuth(authKeys map[string]string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			checkIdentityHeader(c.Request(), utils.GetLogFromEcho(c))

			authorization := c.Request().Header.Get("authorization")

			if authorization == "" {
				return echo.NewHTTPError(http.StatusUnauthorized)
			}

			match := headerMatcher.FindStringSubmatch(authorization)
			if len(match) != 2 {
				return echo.NewHTTPError(http.StatusUnauthorized, "Unsupported authentication key format")
			}

			for principal, key := range authKeys {
				if key == match[1] {
					utils.SetRequestContextValue(c, pskPrincipal, principal)
					return next(c)
				}
			}

			return echo.NewHTTPError(http.StatusForbidden)
		}
	}
}

func GetPSKPrincipal(ctx context.Context) string {
	principal := ctx.Value(pskPrincipal)

	if principal == nil {
		panic("principal not set")
	}

	return principal.(string)
}

func BuildPskAuthConfigFromEnv() map[string]string {
	result := map[string]string{}

	for _, param := range os.Environ() {
		match := envMatcher.FindStringSubmatch(param)

		if len(match) != 3 {
			continue
		}

		principal := strings.ToLower(match[1])
		result[principal] = match[2]
	}

	return result
}

// TODO: enable x509 for auth in the future
func checkIdentityHeader(request *http.Request, log *zap.SugaredLogger) {
	if identity := request.Header.Get("x-rh-identity"); identity != "" {
		log.Debugw("received identity header", "identity", identity)
	}
}
