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
	logs := []models.AlarmLog{}
	// 构建查询条件
	// svc.db.Engine.ShowSQL(true)
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
			err := svc.packALarmLog(&logs[i])
			if err != nil {
				return nil, err
			}
			uniqueLogs = append(uniqueLogs, logs[i])
		}
	}
	return uniqueLogs, nil
}

func (svc *Log) packALarmLog(log *models.AlarmLog) error {
	asset, err := GetAssetByID(svc.db.Engine, log.AssetID)
	if err != nil {
		return err
	}
	log.AssetName = asset.Name
	rule, err := GetRuleByID(svc.db.Engine, log.RuleID)
	if err != nil {
		return err
	}
	log.RuleName = rule.Name
	log.RuleType = rule.Type
	creator, err := GetUserByID(svc.db.Engine, rule.CreatorID)
	if err != nil {
		return err
	}
	log.Admin = creator.Name
	return nil
}

func (svc *Log) FindUserLogs(userID int, conditions map[string]interface{}) ([]models.UserLog, error) {
	logs := []models.UserLog{}
	// 构建查询条件
	// svc.db.Engine.ShowSQL(true)
	queryBuilder := svc.db.Engine.Table("user_log").Alias("log")
	queryBuilder = queryBuilder.Join("LEFT", "user", "log.user_id = user.id")
	queryBuilder = queryBuilder.And("log.user_id = user.id")
	queryBuilder = queryBuilder.Or("user.role >= 30")
	for key, value := range conditions {
		switch key {
		case "username":
			queryBuilder = queryBuilder.And("user.username LIKE ?", "%"+value.(string)+"%")
		case "phone":
			queryBuilder = queryBuilder.And("user.phone LIKE ?", "%"+value.(string)+"%")
		case "module":
			queryBuilder = queryBuilder.And("log.module = ?", value)
		case "type":
			queryBuilder = queryBuilder.And("log.type = ?", value)
		case "ip":
			queryBuilder = queryBuilder.And("log.ip LIKE ?", value.(string)+"%")
		case "createTimeBegin":
			tm := time.Unix(int64(value.(int)), 0).Format("2006-01-02 15:04:05")
			queryBuilder = queryBuilder.And("log.created_at >= ?", tm)
		case "createTimeEnd":
			tm := time.Unix(int64(value.(int)), 0).Format("2006-01-02 15:04:05")
			queryBuilder = queryBuilder.And("log.created_at <= ?", tm)
		}
	}
	_, ok1 := conditions["module"]
	_, ok2 := conditions["type"]
	if !ok1 && !ok2 {
		value := "权限控制"
		queryBuilder = queryBuilder.And("log.module != ?", value)
	}

	err := queryBuilder.Find(&logs)
	if err != nil {
		return nil, err
	}

	processedIDs := make(map[int]bool)
	uniqueLogs := make([]models.UserLog, 0)

	for i := range logs {
		if !processedIDs[logs[i].ID] {
			processedIDs[logs[i].ID] = true
			user, err := GetUserByID(svc.db.Engine, logs[i].UserID)
			if err != nil {
				return nil, err
			}
			if user == nil {
				continue
			}
			err = svc.packUserLog(&logs[i], user)
			if err != nil {
				return nil, err
			}
			uniqueLogs = append(uniqueLogs, logs[i])
		}
	}
	return uniqueLogs, nil
}

func (svc *Log) packUserLog(log *models.UserLog, user *models.UserInfo) error {
	log.Username = user.Username
	log.Phone = user.Phone
	return nil
}

func (svc *Log) GetAlarmLogByID(alarmLogID int) (*models.AlarmLog, error) {
	alarmLog := new(models.AlarmLog)
	has, err := svc.db.Engine.ID(alarmLogID).Get(alarmLog)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	if err := svc.packALarmLog(alarmLog); err != nil {
		return nil, err
	}
	return alarmLog, nil
}

func (svc *Log) GetUserByID(userID int) (*models.UserInfo, error) {
	return GetUserByID(svc.db.Engine, userID)
}
