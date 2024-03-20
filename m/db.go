package models

import (
	"log"

	"xorm.io/xorm"
)

type Database struct {
	engine   *xorm.Engine
}

func NewDatabase(host string, user string, password string) *Database {
	engine, err := xorm.NewEngine("mysql", "")
	if err != nil {
		log.Fatal(err)
	}
	m := &Database{
		engine:   engine,
	}
	return m
}

func (db Database) GetEngine() *xorm.Engine {
	if db.engine == nil {
		log.Fatal("The model's engine has not been initialized")
	}
	return db.engine
}
