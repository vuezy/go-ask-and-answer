package utils

import (
	"fmt"
	"reflect"
	"strings"
)

/*
ConvertStructToMap is meant to be used for converting a struct or a slice of structs
into a map or a slice of maps.
This function should only be used when you are not allowed or do not have access
to add tags to fields of a struct.
Nested slice, nested map, and nested struct are not fully supported.
*/
func ConvertStructToMap(data any) any {
	dataValue := reflect.ValueOf(data)
	if dataValue.Kind() == reflect.Slice {
		newSlice := make([]map[string]any, 0)
		for i := 0; i < dataValue.Len(); i++ {
			newMap := ConvertStructToMap(dataValue.Index(i).Interface()).(map[string]any)
			newSlice = append(newSlice, newMap)
		}
		return newSlice
	}

	structValue := dataValue
	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
	}

	if structValue.Kind() != reflect.Struct {
		fmt.Println("The element is not a struct")
		return map[string]any{}
	}

	newMap := make(map[string]any)
	for i := 0; i < structValue.NumField(); i++ {
		fieldName := structValue.Type().Field(i).Name
		fieldValue := structValue.Field(i)
		newMap[toSnakeCase(fieldName)] = fieldValue.Interface()
	}
	return newMap
}

func toSnakeCase(input string) string {
	var result strings.Builder

	for i, char := range input {
		if i > 0 && char >= 'A' && char <= 'Z' {
			if input[i-1] >= 'a' && input[i-1] <= 'z' {
				result.WriteRune('_')
			} else if i+1 < len(input) && input[i+1] >= 'a' && input[i+1] <= 'z' {
				result.WriteRune('_')
			}
		}
		result.WriteRune(char)
	}

	return strings.ToLower(result.String())
}
