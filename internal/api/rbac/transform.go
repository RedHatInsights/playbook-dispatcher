package rbac

func FilterPermissions(permissions []Access, requiredPermission RequiredPermission) (result []Access) {
	result = []Access{}

	for _, value := range permissions {
		parsed := permissionRegex.FindStringSubmatch(value.Permission)

		if len(parsed) != 4 {
			continue
		}

		if requiredPermission.Application == parsed[1] &&
			matches(requiredPermission.ResourceType, parsed[2]) &&
			matches(requiredPermission.Verb, parsed[3]) {
			result = append(result, value)
		}
	}

	return
}

func GetPredicateValues(permissions []Access, key string) (result []string) {
	for _, permission := range permissions {
		for _, resourceDefinition := range permission.ResourceDefinitions {
			if resourceDefinition.AttributeFilter.Key != key {
				continue
			}

			if resourceDefinition.AttributeFilter.Operation == operationEqual {
				if valueString, ok := resourceDefinition.AttributeFilter.Value.(string); ok {
					result = append(result, valueString)
				} else if valueStringList, ok := resourceDefinition.AttributeFilter.Value.([]string); ok {
					result = append(result, valueStringList...)
				}
			}
		}
	}

	return
}
