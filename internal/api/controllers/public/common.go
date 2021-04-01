package public

import (
	"fmt"
	"math"
	"net/url"
	"playbook-dispatcher/internal/common/utils"
	"strconv"
	"strings"
)

const defaultLimit = 50

// these functions should not be needed - the generated code should fill in default values from the schema
func getLimit(limit *Limit) int {
	if limit != nil {
		return (int(*limit))
	}

	return defaultLimit
}

func getOffset(offset *Offset) int {
	if offset != nil {
		return int(*offset)
	}

	return 0
}

func parseFields(input map[string][]string, key string, knownFields map[string]string, defaults []string) ([]string, error) {
	selectedFields, ok := input[key]

	if !ok {
		return defaults, nil
	}

	result := []string{}

	for _, value := range selectedFields {
		for _, field := range strings.Split(value, ",") {
			if _, ok := knownFields[field]; ok {
				result = append(result, field)
			} else {
				return nil, fmt.Errorf("unknown field: %s", field)
			}
		}
	}

	return result, nil
}

func createLinks(base string, queryString string, limit, offset, total int) Links {
	lastPage := int(math.Floor(float64(utils.Max(total-1, 0)) / float64(limit)))

	links := Links{
		First: createLink(base, queryString, limit, 0),
		Last:  createLink(base, queryString, limit, lastPage*limit),
	}

	if offset > 0 {
		previous := createLink(base, queryString, limit, utils.Max(offset-limit, 0))
		links.Previous = &previous
	}

	if offset+limit < total {
		next := createLink(base, queryString, limit, offset+limit)
		links.Next = &next
	}

	return links
}

func createLink(base string, queryString string, limit, offset int) string {
	query, _ := url.ParseQuery(queryString)

	query.Set("limit", strconv.Itoa(limit))
	query.Set("offset", strconv.Itoa(offset))

	return fmt.Sprintf("%s?%s", base, query.Encode())
}
