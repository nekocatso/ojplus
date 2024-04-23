package models

import (
	"Ojplus/internal/config"

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
