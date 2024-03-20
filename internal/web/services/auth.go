package services

import (
	"Alarm/internal/web/models"
)

type Auth struct {
	db *models.Database
}

func NewAuth(db *models.Database) *Auth {
	return &Auth{db: db}
}

func (svc *Auth) CreateToken() error {
	return nil
}
