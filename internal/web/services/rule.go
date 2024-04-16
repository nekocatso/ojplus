package services

import (
	"Alarm/internal/web/models"
	"errors"
	"fmt"
	"log"
	"time"

	"xorm.io/xorm"
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

func (svc *Rule) CreateRule(rule *models.Rule, pingInfo *models.PingInfo, tcpInfo *models.TCPInfo, assetIDs []int) error {
	session := svc.db.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}

	if rule.AlarmID != 0 {
		has, err := session.ID(rule.AlarmID).Exist(&models.AlarmTemplate{})
		if err != nil {
			return err
		}
		if !has {
			return fmt.Errorf("alarm %d not found", rule.AlarmID)
		}
	}
	_, err = session.Insert(rule)
	if err != nil {
		session.Rollback()
		return err
	}
	if rule.Type == "ping" {
		pingInfo.ID = rule.ID
		_, err = session.Insert(pingInfo)
	} else if rule.Type == "tcp" {
		tcpInfo.ID = rule.ID
		_, err = session.Insert(tcpInfo)
	} else {
		return errors.New("not ping or tcp")
	}
	if err != nil {
		session.Rollback()
		return err
	}
	if len(assetIDs) > 0 {
		err = svc.BindAssets(session, rule.ID, assetIDs)
		if err != nil {
			session.Rollback()
			return err
		}
	}
	return nil
}

func (svc *Rule) UpdateRuleByID(ruleID int, ruleUpdateMap, pingInfoUpdateMap, tcpInfoUpdateMap map[string]interface{}, assetIDs []int) error {
	session := svc.db.Engine.NewSession()
	defer session.Close()

	err := session.Begin()
	if err != nil {
		return err
	}
	rule, err := GetRuleByID(svc.db.Engine, ruleID)
	if err != nil {
		session.Rollback()
		return err
	}
	session.Table(new(models.Rule)).ID(ruleID).Update(ruleUpdateMap)
	if rule.Type == "ping" {
		session.Table(new(models.PingInfo)).ID(ruleID).Update(pingInfoUpdateMap)
	} else if rule.Type == "tcp" {
		session.Table(new(models.TCPInfo)).ID(ruleID).Update(tcpInfoUpdateMap)
	} else {
		return errors.New("not ping or tcp")
	}
	if len(assetIDs) > 0 {
		err = svc.BindAssets(session, rule.ID, assetIDs)
		if err != nil {
			session.Rollback()
			return err
		}
	}
	return nil
}

func (svc *Rule) BindAssets(session *xorm.Session, ruleID int, assetIDs []int) error {
	// 去重
	assetIDsMap := make(map[int]bool)
	for _, assetID := range assetIDs {
		assetIDsMap[assetID] = true
	}
	uniqueAssetIDs := make([]int, 0, len(assetIDsMap))
	for assetID := range assetIDsMap {
		uniqueAssetIDs = append(uniqueAssetIDs, assetID)
	}

	// 删除已有关联
	_, err := session.And("rule_id = ?", ruleID).Delete(&models.AssetRule{})
	if err != nil {
		session.Rollback()
		return err
	}

	// 获取已绑定的资产ID
	var existingAssetIDs []int
	err = session.Table("asset_rule").Cols("asset_id").Where("rule_id = ?", ruleID).Find(&existingAssetIDs)
	if err != nil {
		session.Rollback()
		return err
	}

	// 比较新增的assetIDs和已有的assetIDs，找出新增的和需要删除的
	var toAdd, toDelete []int
	existingAssetIDMap := make(map[int]bool)
	for _, assetID := range existingAssetIDs {
		existingAssetIDMap[assetID] = true
	}
	for _, assetID := range uniqueAssetIDs {
		if !existingAssetIDMap[assetID] {
			toAdd = append(toAdd, assetID)
		}
	}
	for _, assetID := range existingAssetIDs {
		flag := true
		for _, item := range uniqueAssetIDs {
			if item == assetID {
				flag = false
			}
		}
		if flag {
			toDelete = append(toDelete, assetID)
		}
	}

	// 删除多余的关联
	if len(toDelete) > 0 {
		_, err = session.In("asset_id", toDelete).Delete(&models.AssetRule{})
		if err != nil {
			session.Rollback()
			return err
		}
	}

	// 添加新增的关联
	for _, assetID := range toAdd {
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

	return nil
}

func (svc *Rule) GetRuleByID(ruleID, userID int) (*models.Rule, error) {
	rule := new(models.Rule)
	has, err := svc.db.Engine.ID(ruleID).Get(rule)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("rule with ID %d does not exist", ruleID)
	}
	svc.packRule(rule, userID)
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

func (svc *Rule) SetRule(rule *models.Rule) (bool, error) {
	return svc.db.Engine.Get(rule)
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
	rule.AssetsCount, err = svc.GetAssetCount(rule.ID)
	if err != nil {
		return err
	}
	if userID > 0 {
		rule.AssetNames, err = svc.GetAssetNames(rule.ID, userID)
		if err != nil {
			return err
		}
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

func (svc *Rule) GetUserByID(userID int) (*models.User, error) {
	return GetUserByID(svc.db.Engine, userID)
}

func (svc *Rule) DeleteRule(ruleID int) error {
	session := svc.db.Engine.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return err
	}
	_, err = session.ID(ruleID).Delete(new(models.Asset))
	if err != nil {
		session.Rollback()
		return err
	}
	_, err = session.Where("asset_id = ?", ruleID).Delete(new(models.AssetUser))
	if err != nil {
		session.Rollback()
		return err
	}
	var enableAssets []models.AssetRule
	svc.db.Engine.Where("asset_id = ?", ruleID).Join("LEFT", "asset", "asset.id = asset_rule.asset_id").Find(&enableAssets)
	for _, enableAsset := range enableAssets {
		log.Printf("Ctrl Signal: Stop %d\n", enableAsset.ID)
	}
	_, err = session.Where("asset_id = ?", ruleID).Delete(new(models.AssetRule))
	if err != nil {
		session.Rollback()
		return err
	}
	_, err = session.Where("asset_id = ?", ruleID).Delete(new(models.AlarmLog))
	if err != nil {
		session.Rollback()
		return err
	}
	err = session.Commit()
	if err != nil {
		return err
	}
	return nil
}
