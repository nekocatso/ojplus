package services

import (
	"Alarm/internal/web/models"
	"errors"
	"time"
)

type Account struct {
	db    *models.Database
	cache *models.Cache
	cfg   map[string]interface{}
}

func NewAccount(cfg map[string]interface{}) *Account {
	return &Account{
		db:    cfg["db"].(*models.Database),
		cache: cfg["cache"].(*models.Cache),
		cfg:   cfg,
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

func (svc *Account) UpdateUserByID(id int, user *models.User) error {
	updateFields := []string{}
	if user.Password != "" {
		updateFields = append(updateFields, "password")
	}
	if user.Email != "" {
		updateFields = append(updateFields, "email")
	}
	if user.Phone != "" {
		updateFields = append(updateFields, "phone")
	}
	_, err := svc.db.Engine.ID(id).Cols(updateFields...).Update(user)
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

func (svc *Account) GetUserByID(id int) (*models.UserInfo, error) {
	return GetUserByID(svc.db.Engine, id)
}

func (svc *Account) FindUsers(conditions map[string]interface{}) ([]models.UserInfo, error) {
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
	var usersInfo []models.UserInfo
	for _, user := range users {
		usersInfo = append(usersInfo, *user.GetInfo())
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

func (svc *Account) GetUserIDsByAssetID(assetID int) ([]int, error) {
	var userIDs []int
	err := svc.db.Engine.Table("asset_user").Cols("user_id").Where("asset_id =?", assetID).Find(&userIDs)
	if err != nil {
		return nil, err
	}
	return userIDs, nil
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
