package models

import (
	"Alarm/internal/config"
	"reflect"
	"strings"

	"github.com/google/uuid"
	"xorm.io/xorm"
)

type Database struct {
	Engine *xorm.Engine
}

func NewDatabase(cfg *config.Mysql) (*Database, error) {
	engine, err := xorm.NewEngine("mysql", cfg.DSN)
	if err != nil {
		return nil, err
	}
	m := &Database{
		Engine: engine,
	}
	err = engine.Sync(
		new(User),
		new(Asset),
		new(AssetRule),
		new(AssetUser),
		new(Rule),
		new(TCPInfo),
		new(PingInfo),
		new(AlarmTemplate),
		new(AlarmLog),
		new(UserLog),
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

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
