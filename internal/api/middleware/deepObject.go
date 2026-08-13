package middleware

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

func Hack(param string, fields ...string) echo.MiddlewareFunc {
	prefix := buildPrefix(param, fields...)
	escapedPrefix := strings.ReplaceAll(strings.ReplaceAll(prefix, `[`, `\[`), `]`, `\]`)
	regex := regexp.MustCompile(fmt.Sprintf(`^%s\[([^\[\]]+)\]$`, escapedPrefix))

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// TODO: remove parseDeepObject once parseDeepObjectFixed is verified in all environments
			labelFilters := parseDeepObjectFixed(c, c.QueryParams(), regex)
			c.Set(buildKey(prefix), labelFilters)
			return next(c)
		}
	}
}

func parseDeepObject(ctx echo.Context, queryParams url.Values, regex *regexp.Regexp, prefix string) map[string][]string {
	result := make(map[string][]string)

	for key, values := range queryParams {
		matches := regex.FindStringSubmatch(key)

		if len(matches) > 0 {
			for _, value := range values {
				result[matches[1]] = append(result[matches[1]], value)

				requestUrl := ctx.Request().URL
				toRemove := fmt.Sprintf("%s[%s]=%s", prefix, url.PathEscape(matches[1]), value)
				// THIS IS AN UGLY UGLY hack
				// Before you judge me, note that all I really wanted is a simple JSONAPI-like filter
				// using OpenAPI's deepObject style query parameter, e.g.:
				//
				// ?filter[labels][foo]=bar
				//
				// Turns out the golang ecosystem is to mature enough yet to let you do that.
				// I hit roadblocks on multiple levels:
				// - https://github.com/deepmap/oapi-codegen/issues/282
				// - https://github.com/getkin/kin-openapi/issues/293
				// - https://github.com/getkin/kin-openapi/issues/294
				unescapedRaw, _ := url.QueryUnescape(requestUrl.RawQuery)
				replaced := strings.Replace(unescapedRaw, toRemove, "", 1)
				requestUrl.RawQuery = url.QueryEscape(replaced)
				queryParams.Del(key)
			}
		}
	}

	return result
}

// parseDeepObjectFixed fixes a conflict between the Hack middleware and kin-openapi v0.144.0.
// The original parseDeepObject used url.QueryEscape on the entire query string, mangling
// & and = characters. kin-openapi v0.144.0's stricter validation rejects the mangled result.
func parseDeepObjectFixed(ctx echo.Context, queryParams url.Values, regex *regexp.Regexp) map[string][]string {
	result := make(map[string][]string)

	for key, values := range queryParams {
		matches := regex.FindStringSubmatch(key)
		if len(matches) > 0 {
			result[matches[1]] = append(result[matches[1]], values...)
			queryParams.Del(key)
		}
	}
	ctx.Request().URL.RawQuery = queryParams.Encode()

	return result
}

func buildPrefix(param string, fields ...string) string {
	prefix := param

	for _, field := range fields {
		prefix += fmt.Sprintf(`[%s]`, field)
	}

	return prefix
}

func buildKey(prefix string) string {
	return "deepObjectHack|" + prefix
}

func GetDeepObject(ctx echo.Context, param string, fields ...string) map[string][]string {
	if result, ok := ctx.Get(buildKey(buildPrefix(param, fields...))).(map[string][]string); !ok {
		panic("failed to get deep object data")
	} else {
		return result
	}
}

const queryStringKey string = "playbook-dispatcher-query-string"

func CaptureQueryString() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set(queryStringKey, c.QueryString())
			return next(c)
		}
	}
}

func GetQueryString(ctx echo.Context) string {
	return ctx.Get(queryStringKey).(string)
}
