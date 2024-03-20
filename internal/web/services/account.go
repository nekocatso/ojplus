package services

import (
	"Alarm/internal/web/models"
)

type AccountService struct {
	db *models.Database
}

func NewAccount(db *models.Database) *AccountService {
	return &AccountService{db: db}
}

func (svc *AccountService) CreateUser() (*models.User, error) {
	return nil, nil
}
