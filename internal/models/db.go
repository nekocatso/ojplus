package models

import (
	"xorm.io/xorm"
)

type Database struct {
	host     string
	user     string
	password string
	Engine   *xorm.Engine
}

func NewDatabase(host string, user string, password string) (*Database, error) {
	engine, err := xorm.NewEngine("mysql", "")
	if err != nil {
		return nil, err
	}
	m := &Database{
		host:     host,
		user:     user,
		password: password,
		Engine:   engine,
	}
	return m, nil
}
