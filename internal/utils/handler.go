package utils

import (
	"reflect"
	"strings"

	"github.com/google/uuid"
)

func AddUUIDToUniqueFields(data interface{}) {
	v := reflect.ValueOf(data).Elem()
	uuidStr := uuid.New().String()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		tag := field.Tag.Get("xorm")
		if field.Type.Kind() == reflect.String && (strings.Contains(tag, "unique") || strings.Contains(tag, "pk")) {
			fieldValue := v.Field(i)
			if fieldValue.CanSet() {
				fieldValue.SetString(fieldValue.String() + "-" + uuidStr)
			}
		}
	}
}
