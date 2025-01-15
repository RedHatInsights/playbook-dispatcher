package rbac

import "encoding/json"

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

func GetPredicateValues(permissions []Access, key string) (result []string, error error) {
	for _, permission := range permissions {
		for _, resourceDefinition := range permission.ResourceDefinitions {
			if resourceDefinition.AttributeFilter.Key != key {
				continue
			}

			if resourceDefinition.AttributeFilter.Operation == operationEqual {
				// NOTE: This is all super ugly, blame code gen.

				var resourceDefinitionString string
				err := json.Unmarshal(resourceDefinition.AttributeFilter.Value.union, &resourceDefinitionString)
				if err == nil {
					result = append(result, resourceDefinitionString)
					continue
				}

				var resourceDefinitionStringSlice []string
				err = json.Unmarshal(resourceDefinition.AttributeFilter.Value.union, &resourceDefinitionStringSlice)
				if err != nil {
					return result, err
				}
				result = append(result, resourceDefinitionStringSlice...)
			}
		}
	}

	return result, nil
}
