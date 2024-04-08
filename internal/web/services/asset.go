package services

import (
	"Alarm/internal/web/models"
	"errors"
	"fmt"
)

type Asset struct {
	db    *models.Database
	cache *models.Cache
	cfg   map[string]interface{}
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
	if err != nil {
		return err
	}
	return nil
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
	// 解除已有关联
	_, err = session.Where("asset_id = ?", assetID).Delete(&models.AssetRule{})
	if err != nil {
		session.Rollback()
		return err
	}
	// 关联新的规则
	for _, ruleID := range ruleIDs {
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
	return err
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

func (svc *Asset) IsAssetExist(asset *models.Asset) (bool, string, error) {
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

func (svc *Asset) QueryAssetsWithConditions(userID int, conditions map[string]interface{}) ([]models.Asset, error) {
	assets := make([]models.Asset, 0)
	// 构建查询条件
	queryBuilder := svc.db.Engine.Join("INNER", "asset_user", "asset.id = asset_user.asset_id").Where("asset_user.user_id = ?", userID)
	for key, value := range conditions {
		switch key {
		case "name":
			queryBuilder = queryBuilder.And("name = ?", value)
		case "type":
			queryBuilder = queryBuilder.And("type = ?", value)
		case "creatorID":
			queryBuilder = queryBuilder.And("creator_id = ?", value)
		case "createTimeBegin":
			queryBuilder = queryBuilder.And("created_at >= ?", value)
		case "createTimeEnd":
			queryBuilder = queryBuilder.And("created_at <= ?", value)
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

	err := queryBuilder.Find(&assets)
	if err != nil {
		return nil, err
	}
	return assets, nil
}
