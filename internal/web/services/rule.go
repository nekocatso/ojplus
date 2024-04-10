package services

import (
	"Alarm/internal/web/models"
	"errors"
	"fmt"
	"time"
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

func (svc *Rule) BindAssets(ruleID int, assetIDs []int, userID int) error {
	session := svc.db.Engine.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return err
	}
	// 解除已有关联
	_, err = session.And("rule_id = ?", ruleID).Delete(&models.AssetRule{})
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
		return nil, fmt.Errorf("pingInfo with ID %d does not exist", ruleID)
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
		return nil, fmt.Errorf("tcpInfo with ID %d does not exist", ruleID)
	}
	return &tcpInfo, nil
}

func (svc *Rule) FindRules(userID int, conditions map[string]interface{}) ([]models.Rule, error) {
	rules := make([]models.Rule, 0)
	// 查询
	queryBuilder := svc.db.Engine.Table("rule")
	queryBuilder = queryBuilder.Join("LEFT", "asset_rule", "rule.id = asset_rule.rule_id")
	queryBuilder = queryBuilder.Join("LEFT", "asset_user", "asset_rule.asset_id = asset_user.asset_id")
	queryBuilder = queryBuilder.Where("asset_user.user_id = ?", userID)
	queryBuilder = queryBuilder.Or("rule.creator_id = ?", userID)
	for key, value := range conditions {
		switch key {
		case "name":
			queryBuilder = queryBuilder.And("rule.name LIKE ?", "%"+value.(string)+"%")
		case "type":
			queryBuilder = queryBuilder.And("rule.type = ?", value)
		case "creatorID":
			queryBuilder = queryBuilder.And("rule.creator_id = ?", value)
		case "createTimeBegin":
			tm := time.Unix(int64(value.(int)), 0).Format("2006-01-02 15:04:05")
			queryBuilder = queryBuilder.And("rule.created_at >= ?", tm)
		case "createTimeEnd":
			tm := time.Unix(int64(value.(int)), 0).Format("2006-01-02 15:04:05")
			queryBuilder = queryBuilder.And("rule.created_at <= ?", tm)
		case "assetID":
			queryBuilder = queryBuilder.And("asset_rule.asset_id = ?", value)
		}
	}
	err := queryBuilder.Find(&rules)
	if err != nil {
		return nil, err
	}
	for i := range rules {
		rules[i].Creator, err = GetUserByID(svc.db.Engine, rules[i].CreatorID)
		if err != nil {
			return nil, err
		}
		if rules[i].Type == "ping" {
			rules[i].Info, err = svc.GetPingInfo(rules[i].ID)
		} else if rules[i].Type == "tcp" {
			rules[i].Info, err = svc.GetTCPInfo(rules[i].ID)
		}
		if err != nil {
			return nil, err
		}
		rules[i].AssetNames, err = svc.GetAssetNames(rules[i].ID)
		if err != nil {
			return nil, err
		}
		rules[i].AssetsCount, err = svc.GetAssetCount(rules[i].ID)
		if err != nil {
			return nil, err
		}
	}

	return rules, nil
}

func (svc *Rule) GetAssetNames(ruleID int) ([]string, error) {
	assets := []string{}
	err := svc.db.Engine.Table("asset").Join("INNER", "asset_rule", "asset.id = asset_rule.asset_id").Cols("asset.name").Where("asset_rule.rule_id = ?", ruleID).Find(&assets)
	if err != nil {
		return nil, err
	}
	return assets, nil
}

func (svc *Rule) GetAssetCount(ruleID int) (int, error) {
	cnt, err := svc.db.Engine.Where("rule_id = ?", ruleID).Count(&models.AssetRule{})
	return int(cnt), err
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
