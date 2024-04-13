package services

import (
	"Alarm/internal/web/models"

	"time"
)

type Log struct {
	db    *models.Database
	cache *models.Cache
	cfg   map[string]interface{}
}

func NewLog(cfg map[string]interface{}) *Log {
	return &Log{
		db:    cfg["db"].(*models.Database),
		cache: cfg["cache"].(*models.Cache),
		cfg:   cfg,
	}
}

func (svc *Log) FindALarmLogs(userID int, conditions map[string]interface{}) ([]models.AlarmLog, error) {
	logs := make([]models.AlarmLog, 0)
	// 构建查询条件
	queryBuilder := svc.db.Engine.Table("alarm_log").Alias("log")
	queryBuilder = queryBuilder.Join("LEFT", "rule", "log.rule_id = rule.id")
	queryBuilder = queryBuilder.Join("LEFT", "asset", "log.asset_id = asset.id")
	queryBuilder = queryBuilder.Join("LEFT", "asset_user", "log.asset_id = asset.id")
	queryBuilder = queryBuilder.Where("asset_user.user_id = ?", userID).Or("asset.creator_id = ?", userID)
	for key, value := range conditions {
		switch key {
		case "assetID":
			queryBuilder = queryBuilder.And("asset.id = ?", value)
		case "ruleID":
			queryBuilder = queryBuilder.And("rule.id = ?", value)
		case "ruleType":
			queryBuilder = queryBuilder.And("rule.type = ?", value)
		case "state":
			queryBuilder = queryBuilder.And("log.state = ?", value)
		case "assetCreator":
			queryBuilder = queryBuilder.And("asset.creator_id = ?", value)
		case "createTimeBegin":
			tm := time.Unix(int64(value.(int)), 0).Format("2006-01-02 15:04:05")
			queryBuilder = queryBuilder.And("log.created_at >= ?", tm)
		case "createTimeEnd":
			tm := time.Unix(int64(value.(int)), 0).Format("2006-01-02 15:04:05")
			queryBuilder = queryBuilder.And("log.created_at <= ?", tm)
		}
	}
	err := queryBuilder.Find(&logs)
	if err != nil {
		return nil, err
	}

	processedIDs := make(map[int]bool)
	uniqueLogs := make([]models.AlarmLog, 0)

	for i := range logs {
		if !processedIDs[logs[i].ID] {
			processedIDs[logs[i].ID] = true
			asset, err := GetAssetByID(svc.db.Engine, logs[i].AssetID)
			if err != nil {
				return nil, err
			}
			logs[i].AssetName = asset.Name
			rule, err := GetRuleByID(svc.db.Engine, logs[i].RuleID)
			if err != nil {
				return nil, err
			}
			logs[i].RuleName = rule.Name
			uniqueLogs = append(uniqueLogs, logs[i])
		}
	}
	return uniqueLogs, nil
}
