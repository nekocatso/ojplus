package services

import (
	"Alarm/example/models"
	"fmt"
)

type Auth struct {
	db *models.Database
}

func NewAuthService(db *models.Database) *Auth {
	return &Auth{db: db}
}

func (svc *Auth) CreateToken() error {
	var user *models.User
	fmt.Println(user)
	return nil
}
