package services

import (
	"Alarm/internal/web/models"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"xorm.io/xorm"
)

func GetUserByID(engine *xorm.Engine, id int) (*models.UserInfo, error) {
	user := new(models.User)
	has, err := engine.ID(id).Get(user)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return user.GetInfo(), nil
}

func AddUUIDToUniqueFields(data interface{}) {
	v := reflect.ValueOf(data).Elem()
	uuidStr := uuid.New().String()
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		tag := field.Tag.Get("xorm")
		if strings.HasPrefix(tag, "unique") && field.Type.Kind() == reflect.String {
			fieldValue := v.Field(i)
			if fieldValue.CanSet() {
				fieldValue.SetString(fieldValue.String() + "-" + uuidStr)
			}
		}
	}
}
