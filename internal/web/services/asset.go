package services

import (
	"Alarm/internal/pkg/listenerpool"
	"Alarm/internal/web/models"
	"errors"
	"fmt"
	"log"
	"time"

	"xorm.io/xorm"
)

type Asset struct {
	db       *models.Database
	cache    *models.Cache
	listener *listenerpool.ListenerPool
	cfg      map[string]interface{}
}

func NewAsset(cfg map[string]interface{}) *Asset {
	return &Asset{
		db:       cfg["db"].(*models.Database),
		cache:    cfg["cache"].(*models.Cache),
		cfg:      cfg,
		listener: cfg["listener"].(*listenerpool.ListenerPool),
	}
}

func (svc *Asset) SetAsset(asset *models.Asset) (bool, error) {
	return svc.db.Engine.Get(asset)
}

// 添加资产
func (svc *Asset) CreateAsset(asset *models.Asset, userIDs []int, ruleIDs []int) error {
	session := svc.db.Engine.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return err
	}
	_, err = session.Insert(asset)
	if err != nil {
		session.Rollback()
		return err
	}
	err = svc.BindUsers(session, asset.ID, userIDs)
	if err != nil {
		session.Rollback()
		return err
	}
	if len(ruleIDs) > 0 {
		err = svc.BindRules(session, asset.ID, ruleIDs)
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

func (svc *Asset) UpdateAsset(assetID int, updateMap map[string]interface{}, userIDs []int, ruleIDs []int) error {
	session := svc.db.Engine.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return err
	}
	_, err = session.Table(new(models.Asset)).ID(assetID).Update(updateMap)
	if err != nil {
		return err
	}
	err = svc.BindUsers(session, assetID, userIDs)
	if err != nil {
		session.Rollback()
		return err
	}
	err = svc.BindRules(session, assetID, ruleIDs)
	if err != nil {
		session.Rollback()
		return err
	}
	// 提交事务
	err = session.Commit()
	if err != nil {
		session.Rollback()
		return err
	}

	return nil
}

func (svc *Asset) BindUsers(session *xorm.Session, assetID int, userIDs []int) error {
	asset := &models.Asset{}
	has, err := session.ID(assetID).Get(asset)
	if err != nil {
		return err
	}
	if !has {
		return errors.New("asset not found")
	}

	// 创建一个map来存储唯一的userIDs
	userIDsMap := make(map[int]bool)
	userIDsMap[asset.CreatorID] = true
	for _, userID := range userIDs {
		userIDsMap[userID] = true
	}

	existingUserIDs := []int{}
	session.Table("asset_user").Cols("user_id").Where("asset_id = ?", assetID).Find(&existingUserIDs)

	var toAdd, toDelete []int
	existingUserIDMap := make(map[int]bool)

	for _, userID := range existingUserIDs {
		existingUserIDMap[userID] = true
	}

	// 比较新增的userIDs和已有的userIDs，找出新增的和需要删除的
	for userID := range userIDsMap {
		if !existingUserIDMap[userID] {
			toAdd = append(toAdd, userID)
		}
	}

	for _, userID := range existingUserIDs {
		if !userIDsMap[userID] {
			toDelete = append(toDelete, userID)
		}
	}

	// 删除多余的关联
	if len(toDelete) > 0 {
		_, err = session.In("user_id", toDelete).Delete(&models.AssetUser{})
		if err != nil {
			session.Rollback()
			return err
		}
	}

	// 添加新增的关联
	for _, userID := range toAdd {
		exists, err := session.ID(userID).Exist(&models.User{})
		if err != nil {
			session.Rollback()
			return err
		}
		if !exists {
			session.Rollback()
			return fmt.Errorf("user with ID %d does not exist", userID)
		}
		assetUser := &models.AssetUser{
			AssetID: assetID,
			UserID:  userID,
		}
		_, err = session.Insert(assetUser)
		if err != nil {
			session.Rollback()
			return err
		}
	}
	return nil
}

func (svc *Asset) BindRules(session *xorm.Session, assetID int, ruleIDs []int) error {
	// 去重
	ruleIDsMap := make(map[int]bool)
	for _, ruleID := range ruleIDs {
		ruleIDsMap[ruleID] = true
	}
	uniqueRuleIDs := make([]int, 0, len(ruleIDsMap))
	for ruleID := range ruleIDsMap {
		uniqueRuleIDs = append(uniqueRuleIDs, ruleID)
	}
	// 获取当前asset已绑定的rule.id
	var existingRuleIDs []int
	err := session.Table("asset_rule").Cols("rule_id").Where("asset_id = ?", assetID).Find(&existingRuleIDs)
	if err != nil {
		session.Rollback()
		return err
	}
	// 比较新增和已有的ruleIDs，找出新增的和需要删除的
	var toAdd, toDel []int
	existingRuleIDMap := make(map[int]bool)
	for _, ruleID := range existingRuleIDs {
		existingRuleIDMap[ruleID] = true
	}
	for _, ruleID := range uniqueRuleIDs {
		if !existingRuleIDMap[ruleID] {
			toAdd = append(toAdd, ruleID)
		}
	}
	for _, ruleID := range existingRuleIDs {
		flag := true
		for _, item := range uniqueRuleIDs {
			if item == ruleID {
				flag = false
			}
		}
		if flag {
			toDel = append(toDel, ruleID)
		}
	}
	// 删除多余的关联
	if len(toDel) > 0 {
		_, err = session.In("rule_id", toDel).Delete(&models.AssetRule{})
		if err != nil {
			session.Rollback()
			return err
		}
	}
	// 添加新增的关联
	for _, ruleID := range toAdd {
		rule := &models.Rule{ID: ruleID}
		exists, err := session.Exist(rule)
		if err != nil {
			session.Rollback()
			return err
		}
		if !exists {
			session.Rollback()
			return fmt.Errorf("rule with ID %d does not exist", ruleID)
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
		err = svc.listener.AddPing(assetRule.ID)
		if err != nil {
			session.Rollback()
			return err
		}
		log.Printf("Ctrl Signal: Add %d\n", assetRule.ID)
	}
	listenAssetRules := []int{}
	listenAssetRules = append(listenAssetRules, toAdd...)
	listenAssetRules = append(listenAssetRules, toDel...)
	for _, delAssetRule := range listenAssetRules {
		err = svc.listener.Listen(delAssetRule)
		if err != nil {
			session.Rollback()
			return err
		}
		log.Printf("Listen Signal: %d\n", delAssetRule)
	}
	return nil
}

func (svc *Asset) GetAssetInfo(asset *models.Asset) error {
	// 获取资产信息
	_, err := svc.db.Engine.Get(asset)
	if err != nil {
		return err
	}
	return nil
}

func (svc *Asset) GetAssetByID(id int) (*models.Asset, error) {
	asset, err := GetAssetByID(svc.db.Engine, id)
	if err != nil {
		return nil, err
	}
	if asset == nil {
		return nil, errors.New("Asset not found")
	}
	svc.packAsset(asset)
	return asset, nil
}

func (svc *Asset) GetAssetExistInfo(asset *models.Asset) (bool, string, error) {
	has, err := svc.db.Engine.Where("name = ? and creator_id = ?", asset.Name, asset.CreatorID).Exist(&models.Asset{})
	if err != nil {
		return has, "", err
	}
	if has {
		return true, "已存在同名的资产", nil
	}
	has, err = svc.db.Engine.Where("address = ?", asset.Address).Exist(&models.Asset{})
	if err != nil {
		return true, "", err
	}
	if has {
		return true, "已存在同地址的资产", nil
	}
	return false, "", nil
}

func (svc *Asset) IsAssetExistByID(assetID int) (bool, error) {
	return svc.db.Engine.ID(assetID).Exist(&models.Asset{})
}

func (svc *Asset) FindAssets(userID int, conditions map[string]interface{}) ([]models.Asset, error) {
	assets := make([]models.Asset, 0)
	// 构建查询条件
	queryBuilder := svc.db.Engine.Table("asset")
	queryBuilder = queryBuilder.Join("LEFT", "asset_user", "asset.id = asset_user.asset_id")
	queryBuilder = queryBuilder.Join("LEFT", "asset_rule", "asset.id = asset_rule.asset_id")
	queryBuilder = queryBuilder.Join("LEFT", "rule", "rule.id = asset_rule.rule_id")
	for key, value := range conditions {
		switch key {
		case "name":
			queryBuilder = queryBuilder.And("asset.name LIKE ?", "%"+value.(string)+"%")
		case "type":
			queryBuilder = queryBuilder.And("asset.type = ?", value)
		case "creatorID":
			queryBuilder = queryBuilder.And("asset.creator_id = ?", value)
		case "address":
			queryBuilder = queryBuilder.And("asset.address LIKE ?", "%"+value.(string)+"%")
		case "ruleID":
			queryBuilder = queryBuilder.And("asset_rule.rule_id = ?", value)
		case "availableRuleType":
			queryBuilder = queryBuilder.Where(
				`asset.id NOT IN (SELECT asset.id FROM asset LEFT JOIN asset_rule
				ON asset.id = asset_rule.asset_id LEFT JOIN rule
				ON rule.id = asset_rule.rule_id WHERE rule.type = ?)`,
				value)
		case "createTimeBegin":
			tm := time.Unix(int64(value.(int)), 0).Format("2006-01-02 15:04:05")
			queryBuilder = queryBuilder.And("asset.created_at >= ?", tm)
		case "createTimeEnd":
			tm := time.Unix(int64(value.(int)), 0).Format("2006-01-02 15:04:05")
			queryBuilder = queryBuilder.And("asset.created_at <= ?", tm)
		case "enable":
			if value.(int) > 0 {
				queryBuilder = queryBuilder.And("state > 0")
			} else {
				queryBuilder = queryBuilder.And("state = -1")
			}
		case "state":
			queryBuilder = queryBuilder.And("state = ?", value)
		}
	}
	queryBuilder = queryBuilder.And("asset_user.user_id = ? OR asset.creator_id = ?", userID, userID)
	err := queryBuilder.Find(&assets)
	if err != nil {
		return nil, err
	}

	processedIDs := make(map[int]bool)
	uniqueAssets := make([]models.Asset, 0)

	for i := range assets {
		if !processedIDs[assets[i].ID] {
			processedIDs[assets[i].ID] = true
			err := svc.packAsset(&assets[i])
			if err != nil {
				return nil, err
			}
			uniqueAssets = append(uniqueAssets, assets[i])
		}
	}
	return uniqueAssets, nil
}

func (svc *Asset) packAsset(asset *models.Asset) error {
	var err error
	asset.Creator, err = GetUserByID(svc.db.Engine, asset.CreatorID)
	if err != nil {
		return err
	}

	asset.RuleNames, err = svc.GetRuleNames(asset.ID)
	if err != nil {
		fmt.Println(asset.ID)
		return err
	}
	return nil
}

func (svc *Asset) IsAccessAsset(assetID int, userID int) (bool, error) {
	has, err := svc.db.Engine.Where("asset_id = ? AND user_id = ?", assetID, userID).Exist(&models.AssetUser{})
	if err != nil {
		return false, err
	}
	return has, nil
}

func (svc *Asset) GetRuleNames(assetID int) ([]string, error) {
	rules := []models.Rule{}
	err := svc.db.Engine.Table("rule").Join("INNER", "asset_rule", "rule.id = asset_rule.rule_id").Cols("rule.name").Where("asset_rule.asset_id = ?", assetID).Find(&rules)
	if err != nil {
		return nil, err
	}
	ruleNames := []string{}
	for _, rule := range rules {
		ruleNames = append(ruleNames, rule.Name)
	}
	return ruleNames, nil
}

func (svc *Asset) GetUserByID(userID int) (*models.User, error) {
	return GetUserByID(svc.db.Engine, userID)
}

func (svc *Asset) DeleteAsset(assetID int) error {
	session := svc.db.Engine.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return err
	}
	_, err = session.ID(assetID).Delete(new(models.Asset))
	if err != nil {
		session.Rollback()
		return err
	}
	_, err = session.Where("asset_id = ?", assetID).Delete(new(models.AssetUser))
	if err != nil {
		session.Rollback()
		return err
	}
	var enableAssetRules []models.AssetRule
	svc.db.Engine.Where("asset_id = ?", assetID).Join("LEFT", "asset", "asset.id = asset_rule.asset_id").And("asset.state > 0").Find(&enableAssetRules)
	for _, enableAssetRule := range enableAssetRules {
		err := svc.listener.DelPing(enableAssetRule.ID)
		if err != nil {
			session.Rollback()
			return err
		}
		log.Printf("Ctrl Signal: Stop %d\n", enableAssetRule.ID)
	}
	_, err = session.Where("asset_id = ?", assetID).Delete(new(models.AssetRule))
	if err != nil {
		session.Rollback()
		return err
	}
	_, err = session.Where("asset_id = ?", assetID).Delete(new(models.AlarmLog))
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

func (svc *Asset) CountAsset(userID int, conditions map[string]interface{}) (int64, error) {
	assets := new(models.Asset)
	queryBuilder := svc.db.Engine.Table(new(models.Asset))
	queryBuilder = queryBuilder.Join("LEFT", "asset_user", "asset.id = asset_user.asset_id")
	queryBuilder = queryBuilder.Where("asset_user.user_id = ?", userID)
	for key, value := range conditions {
		switch key {
		case "enable":
			if value.(bool) {
				queryBuilder = queryBuilder.And("state > 0")
			} else {
				queryBuilder = queryBuilder.And("state < 0")
			}
		}
	}
	return queryBuilder.Count(&assets)
}
