package models

import (
	"log"

	"xorm.io/xorm"
)

type Database struct {
	host     string
	user     string
	password string
	engine   *xorm.Engine
}

func NewDatabase(host string, user string, password string) *Database {
	engine, err := xorm.NewEngine("mysql", "")
	if err != nil {
		log.Fatal(err)
	}
	m := &Database{
		host:     host,
		user:     user,
		password: password,
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
