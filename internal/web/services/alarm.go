package services

import (
	"Alarm/internal/web/models"
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
	return &alarm, nil
}

func (svc *Alarm) FindAlarms(conditions map[string]interface{}) ([]models.AlarmTemplate, error) {
	var alarms []models.AlarmTemplate
	queryBuilder := svc.db.Engine.Table("user")
	for key, value := range conditions {
		switch key {
		case "username":
			queryBuilder = queryBuilder.And("username LIKE ?", "%"+value.(string)+"%")
		case "name":
			queryBuilder = queryBuilder.And("name LIKE ?", "%"+value.(string)+"%")
		case "phone":
			queryBuilder = queryBuilder.And("phone LIKE ?", value.(string)+"%")
		case "role":
			queryBuilder = queryBuilder.And("role = ?", value)
		case "isActive":
			queryBuilder = queryBuilder.And("is_active = ?", value)
		case "createTimeBegin":
			tm := time.Unix(int64(value.(int)), 0).Format("2006-01-02 15:04:05")
			queryBuilder = queryBuilder.And("user.created_at >= ?", tm)
		case "createTimeEnd":
			tm := time.Unix(int64(value.(int)), 0).Format("2006-01-02 15:04:05")
			queryBuilder = queryBuilder.And("user.created_at <= ?", tm)
		}
	}
	err := queryBuilder.Find(&alarms)
	if err != nil {
		return nil, err
	}
	return alarms, nil
}
