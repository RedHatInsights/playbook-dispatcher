package utils

func MapKeys(value map[string]interface{}) (result []string) {
	for key := range value {
		result = append(result, key)
	}

	return
}

func Min(x, y int) int {
	if x < y {
		return x
	}

	return y
}
