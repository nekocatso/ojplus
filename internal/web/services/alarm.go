package services

import (
	"Alarm/internal/web/models"
	"errors"
	"fmt"
	"time"
)

type Alarm struct {
	db    *models.Database
	cache *models.Cache
	cfg   map[string]interface{}
}

func NewAlarm(cfg map[string]interface{}) *Alarm {
	return &Alarm{
		db:    cfg["db"].(*models.Database),
		cache: cfg["cache"].(*models.Cache),
		cfg:   cfg,
	}
}

func (svc *Alarm) CreateAlarm(alarm *models.AlarmTemplate) error {
	_, err := svc.db.Engine.Insert(alarm)
	return err
}

func (svc *Alarm) SetAlarm(alarm *models.AlarmTemplate) (bool, error) {
	return svc.db.Engine.Get(alarm)
}

func (svc *Alarm) GetAlarmByID(alarmID int) (*models.AlarmTemplate, error) {
	var alarm models.AlarmTemplate
	has, err := svc.db.Engine.ID(alarmID).Get(&alarm)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, fmt.Errorf("alarm with ID %d does not exist", alarmID)
	}
	alarm.RuleNames, err = svc.GetRuleNames(alarm.ID)
	if err != nil {
		return nil, err
	}
	return &alarm, nil
}

func (svc *Alarm) FindAlarms(userID int, conditions map[string]interface{}) ([]models.AlarmTemplate, error) {
	var alarms []models.AlarmTemplate
	queryBuilder := svc.db.Engine.Table("alarm_template").Alias("alarm")
	queryBuilder = queryBuilder.Join("LEFT", "rule", "rule.alarm_id = alarm.id")
	queryBuilder = queryBuilder.Join("LEFT", "asset_rule", "asset_rule.rule_id = rule.id")
	queryBuilder = queryBuilder.Join("LEFT", "asset_user", "asset_rule.asset_id = asset_rule.asset_id")
	queryBuilder = queryBuilder.Where("asset_user.user_id = ?", userID)
	queryBuilder = queryBuilder.Or("alarm.creator_id = ?", userID)
	for key, value := range conditions {
		switch key {
		case "name":
			queryBuilder = queryBuilder.And("alarm.name LIKE ?", "%"+value.(string)+"%")
		case "ruleID":
			queryBuilder = queryBuilder.And("rule.id = ?", value)
		case "createTimeBegin":
			tm := time.Unix(int64(value.(int)), 0).Format("2006-01-02 15:04:05")
			queryBuilder = queryBuilder.And("alarm.created_at >= ?", tm)
		case "createTimeEnd":
			tm := time.Unix(int64(value.(int)), 0).Format("2006-01-02 15:04:05")
			queryBuilder = queryBuilder.And("alarm.created_at <= ?", tm)
		}
	}
	err := queryBuilder.Find(&alarms)
	if err != nil {
		return nil, err
	}
	// 去重
	seen := make(map[int]bool)
	uniqueAlarms := []models.AlarmTemplate{}
	for i := range alarms {
		if !seen[alarms[i].ID] {
			seen[alarms[i].ID] = true
			err := svc.packAlarm(&alarms[i])
			if err != nil {
				return nil, err
			}
			uniqueAlarms = append(uniqueAlarms, alarms[i])
		}
	}
	return uniqueAlarms, nil
}

func (svc *Alarm) packAlarm(alarm *models.AlarmTemplate) error {
	var err error
	alarm.RuleNames, err = svc.GetRuleNames(alarm.ID)
	if err != nil {
		return err
	}
	return nil
}

func (svc *Alarm) DeleteAlarm(alarmID int) error {
	_, err := svc.db.Engine.ID(alarmID).Delete(&models.AlarmTemplate{})
	return err
}

func (svc *Alarm) IsAccessAlarm(alarmID int, userID int) (bool, error) {
	queryBuilder := svc.db.Engine.Table("alarm_template").Alias("alarm")
	queryBuilder = queryBuilder.Join("LEFT", "rule", "rule.alarm_id = alarm.id")
	queryBuilder = queryBuilder.Join("LEFT", "asset_rule", "asset_rule.rule_id = rule.id")
	queryBuilder = queryBuilder.Join("LEFT", "asset_user", "asset_rule.asset_id = asset_rule.asset_id")
	queryBuilder = queryBuilder.Where("asset_user.user_id = ?", userID)
	queryBuilder = queryBuilder.Or("alarm.creator_id = ?", userID)
	return queryBuilder.ID(alarmID).Exist(&models.AlarmTemplate{})
}

func (svc *Alarm) GetRule(alarmID int) (*models.Rule, error) {
	var rule *models.Rule
	has, err := svc.db.Engine.Where("alarm_id = ?", alarmID).Get(&rule)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.New("rule not found")
	}
	return rule, nil
}

func (svc *Alarm) GetRuleNames(alarmID int) ([]string, error) {
	rules := []string{}
	err := svc.db.Engine.Table("rule").Cols("rule.name").Where("alarm_id = ?", alarmID).Find(&rules)
	if err != nil {
		return nil, err
	}
	return rules, nil
}

func (svc *Alarm) GetUserByID(userID int) (*models.User, error) {
	return GetUserByID(svc.db.Engine, userID)
}
