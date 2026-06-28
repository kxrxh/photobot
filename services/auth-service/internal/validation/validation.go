package validation

import (
	"reflect"
	"strings"
)

func GetJSONFieldName(fieldName string, structType reflect.Type) string {
	if structType.Kind() == reflect.Pointer {
		structType = structType.Elem()
	}

	field, found := structType.FieldByName(fieldName)
	if !found {
		return fieldName
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return fieldName
	}

	jsonName := strings.Split(jsonTag, ",")[0]
	if jsonName == "-" {
		return fieldName
	}

	return jsonName
}
