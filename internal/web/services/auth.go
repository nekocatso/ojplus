package services

import (
	"Alarm/internal/web/models"
	"time"
)

type Auth struct {
	db    *models.Database
	cache *models.Cache
}

func NewAuth(db *models.Database, cache *models.Cache) *Auth {
	return &Auth{db: db, cache: cache}
}

func (svc *Auth) RefreshToken(userID int, second int) (string, error) {
	token := ""
	err := svc.cache.Client.Set("key1", "value1", time.Second*time.Duration(second)).Err()
	if err != nil {
		return "", err
	}

	return token, nil
}

func (svc *Auth) DeleteToken() error {
	return nil
}
