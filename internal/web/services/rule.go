package services

import (
	"Alarm/internal/web/models"
	"errors"
	"fmt"
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

func (svc *Rule) CreateRule(rule *models.Rule) error {
	_, err := svc.db.Engine.Insert(rule)
	if err != nil {
		return err
	}
	return nil
}

func (svc *Rule) CreatePingRule(rule *models.Rule, pingInfo *models.PingInfo) error {
	err := svc.CreateRule(rule)
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
	err := svc.CreateRule(rule)
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

func (svc *Rule) BindAssets(ruleID int, assetIDs []int) error {
	session := svc.db.Engine.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return err
	}
	// 解除已有关联
	_, err = session.Where("rule_id = ?", ruleID).Delete(&models.AssetRule{})
	if err != nil {
		session.Rollback()
		return err
	}
	// 关联新的资产
	for _, assetID := range assetIDs {
		asset := &models.Asset{ID: assetID}
		exists, err := session.Exist(asset)
		if err != nil {
			session.Rollback()
			return err
		}
		if !exists {
			session.Rollback()
			return fmt.Errorf("asset with ID %d does not exist", assetID)
		}

		assetRule := &models.AssetRule{
			AssetID: assetID,
			RuleID:  ruleID,
		}
		_, err = session.Insert(assetRule)
		if err != nil {
			session.Rollback()
			return err
		}
	}
	// 提交事务
	err = session.Commit()
	if err != nil {
		session.Rollback()
		return err
	}
	return err
}
