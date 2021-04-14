package utils

func MapKeys(value map[string]interface{}) (result []string) {
	for key := range value {
		result = append(result, key)
	}

	return
}

func MapKeysString(value map[string]string) (result []string) {
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

func Max(x, y int) int {
	if x > y {
		return x
	}

	return x
}

func StringRef(value string) *string {
	return &value
}
