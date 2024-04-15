package services

import (
	"Alarm/internal/pkg/listenerpool"
	"Alarm/internal/web/models"
	"errors"
	"fmt"
	"time"
)

type Asset struct {
	db       *models.Database
	cache    *models.Cache
	listener *listenerpool.ListenerPool
	cfg      map[string]interface{}
}

func NewAsset(cfg map[string]interface{}) *Asset {
	return &Asset{
		db:    cfg["db"].(*models.Database),
		cache: cfg["cache"].(*models.Cache),
		cfg:   cfg,
	}
}

// 添加资产
func (svc *Asset) CreateAsset(asset *models.Asset) error {
	// 在数据库中插入资产
	_, err := svc.db.Engine.Insert(asset)
	return err
}

func (svc *Asset) BindUsers(assetID int, userIDs []int) error {
	session := svc.db.Engine.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return err
	}
	// 解除已有关联
	_, err = session.Where("asset_id = ?", assetID).Delete(&models.AssetUser{})
	if err != nil {
		session.Rollback()
		return err
	}
	// 关联新的用户
	for _, userID := range userIDs {
		user := &models.User{ID: userID}
		exists, err := session.Exist(user)
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
	// 提交事务
	err = session.Commit()
	if err != nil {
		session.Rollback()
		return err
	}
	return err
}

func (svc *Asset) BindRules(assetID int, ruleIDs []int) error {
	session := svc.db.Engine.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return err
	}
	// 获取当前asset已绑定的rule.id
	var existingRuleIDs []int
	err = session.Table("asset_rule").Cols("rule_id").Where("asset_id = ?", assetID).Find(&existingRuleIDs)
	if err != nil {
		session.Rollback()
		return err
	}
	// 比较新增和已有的ruleIDs，找出新增的和需要删除的
	var toAdd, toDelete []int
	existingRuleIDMap := make(map[int]bool)
	for _, ruleID := range existingRuleIDs {
		existingRuleIDMap[ruleID] = true
	}
	for _, ruleID := range ruleIDs {
		if !existingRuleIDMap[ruleID] {
			toAdd = append(toAdd, ruleID)
		}
	}
	for _, ruleID := range existingRuleIDs {
		flag := true
		for _, item := range ruleIDs {
			if item == ruleID {
				flag = false
			}
		}
		if flag {
			toDelete = append(toDelete, ruleID)
		}
	}

	// 删除多余的关联
	if len(toDelete) > 0 {
		_, err = session.In("rule_id", toDelete).Delete(&models.AssetRule{})
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
	}

	// 提交事务
	err = session.Commit()
	if err != nil {
		session.Rollback()
		return err
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

// 根据 ID 获取资产
func (svc *Asset) GetAssetByID(id int) (*models.Asset, error) {
	asset := &models.Asset{}
	// 从数据库中根据 ID 查询资产
	has, err := svc.db.Engine.ID(id).Get(asset)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Asset not found")
	}
	asset.RuleNames, err = svc.GetRuleNames(asset.ID)
	if err != nil {
		return nil, err
	}
	return asset, nil
}

// 更新资产
func (svc *Asset) UpdateAsset(asset *models.Asset) error {
	// 更新数据库中的资产信息
	_, err := svc.db.Engine.ID(asset.ID).Update(asset)
	if err != nil {
		return err
	}
	return nil
}

// 删除资产
func (svc *Asset) DeleteAsset(asset *models.Asset) error {
	// 逻辑删除资产
	_, err := svc.db.Engine.ID(asset.ID).Delete()
	if err != nil {
		return err
	}
	return nil
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
			// 标记ID为已处理
			processedIDs[assets[i].ID] = true

			assets[i].Creator, err = GetUserByID(svc.db.Engine, assets[i].CreatorID)
			if err != nil {
				return nil, err
			}

			assets[i].RuleNames, err = svc.GetRuleNames(assets[i].ID)
			if err != nil {
				fmt.Println(assets[i].ID)
				return nil, err
			}

			// 添加到去重后的列表中
			uniqueAssets = append(uniqueAssets, assets[i])
		}
	}
	return uniqueAssets, nil
}

func (svc *Asset) IsAccessAsset(assetID int, userID int) (bool, error) {
	has, err := svc.db.Engine.Where("asset_id = ? AND user_id = ?", assetID, userID).Exist(&models.AssetUser{})
	if err != nil {
		return false, err
	}
	return has, nil
}

func (svc *Asset) GetRuleNames(assetID int) ([]string, error) {
	rules := []string{}
	err := svc.db.Engine.Table("rule").Join("INNER", "asset_rule", "rule.id = asset_rule.rule_id").Cols("rule.name").Where("asset_rule.asset_id = ?", assetID).Find(&rules)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (svc *Asset) GetAssetsByRuleID(ruleID int) ([]*models.Asset, error) {
	assets := []*models.Asset{}
	err := svc.db.Engine.Table("asset").Join("INNER", "asset_rule", "asset.id = asset_rule.asset_id").Where("asset_rule.rule_id = ?", ruleID).Find(&assets)
	if err != nil {
		return nil, err
	}
	for i := range assets {
		assets[i].Creator, err = GetUserByID(svc.db.Engine, assets[i].CreatorID)
		if err != nil {
			return nil, err
		}
		assets[i].RuleNames, err = svc.GetRuleNames(assets[i].ID)
		if err != nil {
			return nil, err
		}
	}
	return assets, nil
}
