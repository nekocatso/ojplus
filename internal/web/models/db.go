package models

import (
	"Alarm/internal/config"

	"xorm.io/xorm"
)

type Database struct {
	host     string
	user     string
	password string
	Engine   *xorm.Engine
}

func NewDatabase(cfg *config.MysqlConfig) (*Database, error) {
	engine, err := xorm.NewEngine("mysql", "")
	if err != nil {
		return nil, err
	}
	m := &Database{
		host:     cfg.Host,
		user:     cfg.User,
		password: cfg.Password,
		Engine:   engine,
	}
	return m, nil
}
