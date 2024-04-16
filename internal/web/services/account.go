package services

import (
	"Alarm/internal/config"
	"Alarm/internal/web/models"
	"errors"
	"time"
)

type Account struct {
	db     *models.Database
	cache  *models.Cache
	cfg    map[string]interface{}
	global *config.Global
}

func NewAccount(cfg map[string]interface{}) *Account {
	return &Account{
		db:     cfg["db"].(*models.Database),
		cache:  cfg["cache"].(*models.Cache),
		cfg:    cfg,
		global: cfg["global"].(*config.Global),
	}
}

func (svc *Account) CreateUser(user *models.User) error {
	_, err := svc.db.Engine.Insert(user)
	return err
}

func (svc *Account) DeleteUser(user *models.User) error {
	if user.ID == 0 {
		has, err := svc.getUser(user)
		if err != nil {
			return err
		}
		if !has {
			return errors.New("无法找到user数据")
		}
	}
	_, err := svc.db.Engine.ID(user.ID).Delete(user)
	return err
}

func (svc *Account) UpdateUserByID(userID int, updateMap map[string]interface{}) error {
	_, err := svc.db.Engine.Table(new(models.User)).ID(userID).Update(updateMap)
	return err
}

func (svc *Account) RestPassword(userID int) error {
	password := svc.global.Gin.Account.DefaultPassword
	_, err := svc.db.Engine.Table(new(models.User)).ID(userID).Update(map[string]interface{}{"password": password})
	return err
}

func (svc *Account) getUser(user *models.User) (bool, error) {
	if user.ID != 0 {
		has, err := svc.db.Engine.ID(user.ID).Get(user)
		if err != nil {
			return false, err
		}
		if !has {
			return false, nil
		}
		return true, nil
	}
	return svc.db.Engine.Cols("id", "username", "email", "phone").Get(user)
}

func (svc *Account) GetUserByID(id int) (*models.User, error) {
	return GetUserByID(svc.db.Engine, id)
}

func (svc *Account) FindUsers(conditions map[string]interface{}) ([]models.User, error) {
	var users []models.User
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
	err := queryBuilder.Find(&users)
	if err != nil {
		return nil, err
	}
	var usersInfo []models.User
	for _, user := range users {
		usersInfo = append(usersInfo, user)
	}
	return usersInfo, nil
}

func (svc *Account) GetUserExistInfo(user *models.User) (bool, string, error) {
	has, err := svc.db.Engine.Where("username = ?", user.Username).Exist(&models.User{})
	if err != nil {
		return has, "", err
	}
	if has {
		return true, "该用户名已被使用", nil
	}
	if user.Email != "" {
		has, err = svc.db.Engine.Where("email = ?", user.Email).Exist(&models.User{})
		if err != nil {
			return true, "", err
		}
		if has {
			return true, "该邮箱已被使用", nil
		}
	}
	if user.Phone != "" {
		has, err = svc.db.Engine.Where("phone = ?", user.Phone).Exist(&models.User{})
		if err != nil {
			return true, "", err
		}
		if has {
			return true, "该电话号码已被使用", nil
		}
	}
	return false, "", nil
}

func (svc *Account) IsUserIDExist(userID int) (bool, error) {
	return svc.db.Engine.ID(userID).Exist(&models.User{})
}

func (svc *Account) GetUsersByAssetID(assetID int) ([]models.User, error) {
	var users []models.User
	err := svc.db.Engine.Join("LEFT", "asset_user", "asset_user.user_id = user.id").Where("asset_user.asset_id = ?", assetID).Find(&users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (svc *Account) VerifyPassword(username string, password string) (bool, error) {
	user := &models.User{
		Username: username,
		Password: password,
	}
	has, err := svc.db.Engine.Exist(user)
	if err != nil {
		return false, err
	}
	return has, nil
}

func (svc *Account) DeleteUserByID(userID int) error {
	session := svc.db.Engine.NewSession()
	defer session.Close()
	err := session.Begin()
	if err != nil {
		return err
	}
	_, err = session.ID(userID).Delete(new(models.User))
	if err != nil {
		session.Rollback()
		return err
	}
	superAdminID := svc.global.Gin.Account.SuperAdminID
	_, err = session.Table("asset_user").Where("user_id = ?", userID).Update(map[string]int{"user_id": superAdminID})
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
