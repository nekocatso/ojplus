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
	if rule.AlarmID != 0 {
		has, err := svc.db.Engine.ID(rule.AlarmID).Exist(&models.AlarmTemplate{})
		if err != nil {
			return err
		}
		if !has {
			return fmt.Errorf("alarm %d not found", rule.AlarmID)
		}
	}
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
		return errors.New("rule create error")
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
		return errors.New("rule create error")
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

func (svc *Rule) GetPingInfo(ruleID int) (*models.PingInfo, error) {
	var pingInfo models.PingInfo
	exists, err := svc.db.Engine.ID(ruleID).Get(&pingInfo)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("rule with ID %d does not exist", ruleID)
	}
	return &pingInfo, nil
}

func (svc *Rule) GetTCPInfo(ruleID int) (*models.TCPInfo, error) {
	var tcpInfo models.TCPInfo
	exists, err := svc.db.Engine.ID(ruleID).Get(&tcpInfo)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("rule with ID %d does not exist", ruleID)
	}
	return &tcpInfo, nil
}

func (svc *Rule) FindRules(userID int) ([]models.Rule, error) {
	rules := make([]models.Rule, 0)
	// 查询
	queryBuilder := svc.db.Engine.Join("INNER", "asset_rule", "rule.id = asset_rule.rule_id")
	queryBuilder = queryBuilder.Join("INNER", "asset_user", "asset_rule.asset_id = asset_user.asset_id")
	queryBuilder = queryBuilder.Where("asset_user.user_id = ?", userID)
	err := queryBuilder.Find(&rules)
	if err != nil {
		return nil, err
	}
	// 去重
	seen := make(map[int]bool)
	uniqueRules := make([]models.Rule, 0)
	for _, rule := range rules {
		if !seen[rule.ID] {
			seen[rule.ID] = true
			uniqueRules = append(uniqueRules, rule)
		}
	}
	for i := range uniqueRules {
		uniqueRules[i].Creator, err = GetUserByID(svc.db.Engine, uniqueRules[i].CreatorID)
		if err != nil {
			return nil, err
		}
		if uniqueRules[i].Type == "ping" {
			uniqueRules[i].Info, err = svc.GetPingInfo(uniqueRules[i].ID)
		} else if uniqueRules[i].Type == "tcp" {
			uniqueRules[i].Info, err = svc.GetTCPInfo(uniqueRules[i].ID)
		}
		if err != nil {
			return nil, err
		}
	}

	return uniqueRules, nil
}

func (svc *Rule) GetRuleIDsByAssetID(assetID int) ([]int, error) {
	var ruleIDs []int
	err := svc.db.Engine.Table("asset_rule").Where("asset_id = ?", assetID).Cols("rule_id").Find(&ruleIDs)
	if err != nil {
		return nil, err
	}
	return ruleIDs, nil
}

func (svc *Rule) IsAccessRule(ruleID, userID int) (bool, error) {
	queryBuilder := svc.db.Engine.Join("INNER", "asset_rule", "rule.id = asset_rule.rule_id")
	queryBuilder = queryBuilder.Join("INNER", "asset_user", "asset_rule.asset_id = asset_user.asset_id")
	queryBuilder = queryBuilder.Where("asset_user.user_id = ?", userID)
	has, err := queryBuilder.ID(ruleID).Exist(&models.Rule{})
	if err != nil {
		return false, err
	}
	return has, nil
}

func (svc *Rule) IsAccessAsset(assetID int, userID int) (bool, error) {
	has, err := svc.db.Engine.Where("asset_id = ? AND user_id = ?", assetID, userID).Exist(&models.AssetUser{})
	if err != nil {
		return false, err
	}
	return has, nil
}
