package utils

func MapKeys(value map[string]interface{}) (result []string) {
	for key := range value {
		result = append(result, key)
	}

	return
}
