package controllers

import (
	"fmt"
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
