package services

import (
	"Alarm/internal/web/models"
	"errors"
	"time"
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
func (a *Asset) CreateAsset(asset *models.Asset) error {
	// 在数据库中插入资产
	_, err := a.db.Engine.Insert(asset)
	if err != nil {
		return err
	}
	return nil
}

// 根据 ID 获取资产
func (a *Asset) GetAssetByID(id int) (*models.Asset, error) {
	asset := &models.Asset{}
	// 从数据库中根据 ID 查询资产
	has, err := a.db.Engine.ID(id).Get(asset)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("Asset not found")
	}
	return asset, nil
}

// 更新资产
func (a *Asset) UpdateAsset(asset *models.Asset) error {
	// 更新数据库中的资产信息
	_, err := a.db.Engine.ID(asset.ID).Update(asset)
	if err != nil {
		return err
	}
	return nil
}

// 删除资产
func (a *Asset) DeleteAsset(asset *models.Asset) error {
	// 逻辑删除资产：设置 DeletedAt 字段为当前时间
	asset.DeletedAt = time.Now()
	_, err := a.db.Engine.ID(asset.ID).Cols("deleted_at").Update(asset)
	if err != nil {
		return err
	}
	return nil
}
