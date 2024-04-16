package services

import (
	"Alarm/internal/web/models"

	"xorm.io/xorm"
)

func GetUserByID(engine *xorm.Engine, id int) (*models.User, error) {
	user := new(models.User)
	has, err := engine.ID(id).Get(user)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return user, nil
}

func GetAssetByID(engine *xorm.Engine, id int) (*models.Asset, error) {
	asset := new(models.Asset)
	has, err := engine.ID(id).Get(asset)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return asset, nil
}

func GetRuleByID(engine *xorm.Engine, id int) (*models.Rule, error) {
	rule := new(models.Rule)
	has, err := engine.ID(id).Get(rule)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return rule, nil
}

func GetAlarmTemplateByID(engine *xorm.Engine, id int) (*models.AlarmTemplate, error) {
	alarmTemplate := new(models.AlarmTemplate)
	has, err := engine.ID(id).Get(alarmTemplate)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return alarmTemplate, nil
}
