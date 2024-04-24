package services

import "Ojplus/internal/web/models"

func (svc *Auth) GetUserExistInfo(username, email string) (string, error) {
	if username != "" {
		exist, err := svc.db.Engine.Where("username = ?", username).Exist(&models.User{})
		if err != nil {
			return "", err
		}
		if exist {
			return "该学号已被注册", nil
		}
	}
	if email != "" {
		exist, err := svc.db.Engine.Where("email = ?", email).Exist(&models.User{})
		if err != nil {
			return "", err
		}
		if exist {
			return "该邮箱已被注册", nil
		}
	}
	return "", nil
}

func (svc *Auth) CreateUser(form any) (int, error) {
	userManager := models.NewUserManager(svc.db)
	return userManager.Create(form)
}

func (svc *Auth) UpdateUser(userID int, form any) error {
	userManager := models.NewUserManager(svc.db)
	return userManager.Update(userID, form)
}

func (svc *Auth) DeleteUser(userID int) error {
	userManager := models.NewUserManager(svc.db)
	return userManager.Delete(userID)
}

func (svc *Auth) GetUserByID(userID int) (*models.User, error) {
	userManager := models.NewUserManager(svc.db)
	return userManager.GetByID(userID)
}

// 按页查询，返回用户列表和总数
func (svc *Auth) GetUsersByPage(page, pageSize int) ([]models.User, int64, error) {
	userManager := models.NewUserManager(svc.db)
	users, err := userManager.GetByPage(page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if users == nil {
		return []models.User{}, 0, nil
	}
	cnt, err := userManager.Count("")
	if err != nil {
		return nil, 0, err
	}
	return users, cnt, nil
}

func (svc *Auth) GetUserIDByAccount(account string) (int, error) {
	var userID int
	exist, err := svc.db.Engine.Table(new(models.User)).Where("username = ?", account).Or("email = ?", account).Cols("id").Get(&userID)
	if err != nil {
		return 0, err
	}
	if !exist {
		return 0, nil
	}
	return userID, nil
}
