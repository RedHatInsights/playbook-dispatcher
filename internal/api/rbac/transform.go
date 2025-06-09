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

func GetPredicateValues(permissions []Access, key string) (result []string) {
	for _, permission := range permissions {
		for _, resourceDefinition := range permission.ResourceDefinitions {

			var operationEqual ResourceDefinitionFilterOperationEqual
			err := json.Unmarshal(resourceDefinition.AttributeFilter.union, &operationEqual)
			if err == nil {
				if operationEqual.Key != key {
					continue
				}

				if operationEqual.Value != nil {
					result = append(result, *operationEqual.Value)
					continue
				}
			}

			var operationIn ResourceDefinitionFilterOperationIn
			err = json.Unmarshal(resourceDefinition.AttributeFilter.union, &operationIn)
			if err == nil {
				if operationIn.Key != key {
					continue
				}

				result = append(result, operationIn.Value...)
			}
		}
	}

	return
}
