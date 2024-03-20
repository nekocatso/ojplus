package services

import (
	"Alarm/internal/web/models"
)

type AuthService struct {
	db *models.Database
}

func NewAuth(db *models.Database) *AuthService {
	return &AuthService{db: db}
}

func (svc *AuthService) CreateToken() error {
	return nil
}
