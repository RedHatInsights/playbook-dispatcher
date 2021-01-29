package middleware

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/labstack/echo/v4"
)

const labelFiltersKey = "playbook-dispatcher-runs-label-filters"

var regex = regexp.MustCompile(`^filter\[labels\]\[([^\[\]]+)\]$`)

func LabelFiltersHack(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		labelFilters := parseLabelFilters(c, c.QueryParams(), true)
		c.Set(labelFiltersKey, labelFilters)
		return next(c)
	}
}

func parseLabelFilters(ctx echo.Context, queryParams url.Values, mutate bool) map[string]string {
	labels := make(map[string]string)

	for key, values := range queryParams {
		matches := regex.FindStringSubmatch(key)

		if len(matches) > 0 {
			for _, value := range values {
				labels[matches[1]] = value

				if mutate {
					requestUrl := ctx.Request().URL
					toRemove := fmt.Sprintf("filter[labels][%s]=%s", url.PathEscape(matches[1]), url.PathEscape(value))
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
					requestUrl.RawQuery = strings.Replace(requestUrl.RawQuery, toRemove, "", 1)
					queryParams.Del(key)
				}
			}
		}
	}

	return labels
}

func GetLabelFilters(ctx echo.Context) map[string]string {
	if result, ok := ctx.Get(labelFiltersKey).(map[string]string); !ok {
		panic("failed to get label filters")
	} else {
		return result
	}
}
