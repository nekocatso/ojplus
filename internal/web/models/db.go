package models

import (
	"Ojplus/internal/config"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

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
	)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func addDelSuffix(model any) error {
	val := reflect.ValueOf(model)
	suffix := fmt.Sprintf("-del-%s", time.Now().Format("20060102150405"))
	if val.Kind() != reflect.Ptr {
		return errors.New("model not a pointer")
	}
	val = val.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("model not a struct")
	}
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if field.Kind() == reflect.String {
			tag := val.Type().Field(i).Tag.Get("xorm")
			if strings.Contains(tag, "unique") {
				field.SetString(field.String() + suffix)
			}
		}
	}
	return nil
}
