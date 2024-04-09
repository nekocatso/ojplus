package services

import (
	"Alarm/internal/web/models"
	"errors"
)

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

func (svc *Rule) CreatePingRule(rule *models.Rule, pingInfo *models.PingInfo) error {
	_, err := svc.db.Engine.Insert(rule)
	if err != nil {
		return err
	}
	has, err := svc.SetRule(rule)
	if err != nil {
		return err
	}
	if !has || rule.ID == 0 {
		return errors.New("rule not found")
	}
	pingInfo.ID = rule.ID
	_, err = svc.db.Engine.Insert(pingInfo)
	if err != nil {
		return err
	}
	return err
}
func (svc *Rule) CreateTCPRule(rule *models.Rule, tcpInfo *models.TCPInfo) error {
	_, err := svc.db.Engine.Insert(rule)
	if err != nil {
		return err
	}
	has, err := svc.SetRule(rule)
	if err != nil {
		return err
	}
	if !has || rule.ID == 0 {
		return errors.New("rule not found")
	}
	tcpInfo.ID = rule.ID
	_, err = svc.db.Engine.Insert(tcpInfo)
	if err != nil {
		return err
	}
	return err
}

func (svc *Rule) SetRule(rule *models.Rule) (bool, error) {
	return svc.db.Engine.Get(rule)
}
