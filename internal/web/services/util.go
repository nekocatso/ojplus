package services

import (
	"Alarm/internal/web/models"

	"xorm.io/xorm"
)

func GetUserByID(engine *xorm.Engine, id int) (*models.UserInfo, error) {
	user := new(models.User)
	has, err := engine.ID(id).Get(user)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return user.GetInfo(), nil
}
