package services

import (
	"Alarm/example/models"
	"fmt"
)

type AuthService struct {
	db *models.Database 
}

func NewAuthService(db *models.Database) *AuthService {
	return &AuthService{db: db}
}

func (svc *AuthService) CreateToken() error {
	var user *models.User
	fmt.Println(user)
	return nil
}
