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

func (svc *Rule) GetRuleByID(ruleID int) (*models.Rule, error) {
	rule := new(models.Rule)
	has, err := svc.db.Engine.ID(ruleID).Get(rule)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("rule with ID %d does not exist", ruleID)
	}
	return rule, nil
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
	queryBuilder = queryBuilder.And("asset_user.user_id = ?", userID)
	queryBuilder = queryBuilder.Or("rule.creator_id = ?", userID)
	for key, value := range conditions {
		switch key {
		case "name":
			queryBuilder = queryBuilder.And("name LIKE ?", "%"+value.(string)+"%")
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
		case "alarmID":
			queryBuilder = queryBuilder.And("rule.alarm_id = ?", value)
		}
	}
	err := queryBuilder.Find(&rules)
	if err != nil {
		return nil, err
	}

	processedIDs := make(map[int]bool)
	uniqueRules := make([]models.Rule, 0)
	// 处理响应内容
	for i := range rules {
		if !processedIDs[rules[i].ID] {
			processedIDs[rules[i].ID] = true
			err := svc.packRule(&rules[i], userID)
			if err != nil {
				return nil, err
			}
			uniqueRules = append(uniqueRules, rules[i])
		}
	}
	return uniqueRules, nil
}

func (svc *Rule) packRule(rule *models.Rule, userID int) error {
	var err error
	rule.Creator, err = GetUserByID(svc.db.Engine, rule.CreatorID)
	if err != nil {
		return err
	}
	if rule.Type == "ping" {
		rule.Info, err = svc.GetPingInfo(rule.ID)
	} else if rule.Type == "tcp" {
		rule.Info, err = svc.GetTCPInfo(rule.ID)
	}

	if err != nil || rule.Info == nil {
		return err
	}
	rule.AssetNames, err = svc.GetAssetNames(rule.ID, userID)
	if err != nil {
		return err
	}
	rule.AssetsCount, err = svc.GetAssetCount(rule.ID)
	if err != nil {
		return err
	}
	return nil
}

func (svc *Rule) GetAssetNames(ruleID, userID int) ([]string, error) {
	assets := []string{}
	assetsMap := map[string]bool{}
	queryBuilder := svc.db.Engine.Table("asset")
	queryBuilder = queryBuilder.Join("LEFT", "asset_rule", "asset.id = asset_rule.asset_id")
	queryBuilder = queryBuilder.Join("LEFT", "asset_user", "asset.id = asset_user.asset_id")
	queryBuilder = queryBuilder.Where("asset_user.user_id = ?", userID).Or("asset.creator_id = ?", userID)
	err := queryBuilder.Cols("asset.name").Where("asset_rule.rule_id = ?", ruleID).Find(&assets)
	if err != nil {
		return nil, err
	}
	for _, asset := range assets {
		assetsMap[asset] = true
	}
	uniqueAssets := []string{}
	for asset := range assetsMap {
		uniqueAssets = append(uniqueAssets, asset)
	}
	return uniqueAssets, nil
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
	queryBuilder := svc.db.Engine.Join("LEFT", "asset_rule", "rule.id = asset_rule.rule_id")
	queryBuilder = queryBuilder.Join("LEFT", "asset_user", "asset_rule.asset_id = asset_user.asset_id")
	queryBuilder = queryBuilder.Where("asset_user.user_id = ?", userID)
	queryBuilder = queryBuilder.Or("rule.creator_id = ?", userID)
	return queryBuilder.ID(ruleID).Exist(&models.Rule{})
}

func (svc *Rule) IsAccessAsset(assetID int, userID int) (bool, error) {
	return svc.db.Engine.Where("asset_id = ? AND user_id = ?", assetID, userID).Exist(&models.AssetUser{})
}

func (svc *Rule) GetUserByID(userID int) (*models.UserInfo, error) {
	return GetUserByID(svc.db.Engine, userID)
}
