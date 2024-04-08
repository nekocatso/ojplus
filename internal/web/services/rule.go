package services

import "Alarm/internal/web/models"

type Rule struct {
	db    *models.Database
	cache *models.Cache
	cfg   map[string]interface{}
}

func NewRule(cfg map[string]interface{}) *Rule {
	return &Rule{
		db:    cfg["db"].(*models.Database),
		cache: cfg["cache"].(*models.Cache),
		cfg:   cfg,
	}
}
