package services

import (
	"Ojplus/internal/web/models"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Auth struct {
	db         *models.Database
	cache      *models.Cache
	privateKey any
	publicKey  any
	cfg        map[string]any
}

func NewAuth(cfg map[string]any) *Auth {
	return &Auth{
		db:         cfg["db"].(*models.Database),
		cache:      cfg["cache"].(*models.Cache),
		privateKey: cfg["privateKey"],
		publicKey:  cfg["publicKey"],
		cfg:        cfg,
	}
}

// User
func (svc *Auth) GetUserExistInfo(username, email string) (string, error) {
	exist, err := svc.db.Engine.Where("username = ?", username).Exist(&models.User{})
	if err != nil {
		return "", err
	}
	if exist {
		return "该学号已被注册", nil
	}
	exist, err = svc.db.Engine.Where("email = ?", username).Exist(&models.User{})
	if err != nil {
		return "", err
	}
	if exist {
		return "该邮箱已被注册", nil
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

func (svc *Auth) GenerateToken(userID int) (map[string]string, error) {
	accessToken, err := svc.generateOneToken(userID, "access")
	if err != nil {
		return nil, err
	}
	refreshToken, err := svc.generateOneToken(userID, "refresh")
	if err != nil {
		return nil, err
	}
	data := map[string]string{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}
	return data, nil
}

func (svc *Auth) generateOneToken(userID int, tokenType string) (string, error) {
	validSeconds := svc.cfg["refreshTokenValidity"].(int)
	claims := jwt.MapClaims{
		"exp":    time.Now().Add(time.Second * time.Duration(validSeconds)).Unix(),
		"userID": userID,
		"type":   tokenType,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenStr, err := token.SignedString(svc.privateKey)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (svc *Auth) ParseToken(tokenStr string) (map[string]any, error) {
	// 解析令牌
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		// 验证签名方法是否为RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("invalid signing method: %v", token.Header["alg"])
		}
		return svc.publicKey, nil
	})
	if err != nil {
		return nil, err
	}
	// 验证令牌是否有效
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	// 解析声明
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	// 验证过期时间
	expirationTime := time.Unix(int64(claims["exp"].(float64)), 0)
	if time.Now().UTC().After(expirationTime) {
		return nil, fmt.Errorf("token has expired")
	}
	return map[string]any(claims), nil
}

func (svc *Auth) CheckPermisson(userID int, role int) (bool, error) {
	if userID <= 0 {
		return false, nil
	}
	if role <= 0 {
		return true, nil
	}
	userManager := models.NewUserManager(svc.db)
	user, err := userManager.GetByID(userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, errors.New("user not found")
	}
	if !*user.Disable {
		return false, nil
	}
	return *user.Role >= role, nil
}

// 校验账号密码，成功返回UserID，失败返回0
func (svc *Auth) VerifyPassword(userID int, password string) (bool, error) {
	if userID <= 0 || password == "" {
		return false, nil
	}
	return svc.db.Engine.ID(userID).And("password = ?", password).Exist()
}

func (svc *Auth) GetUserIDByAccount(account string) (int, error) {
	var userID int
	exist, err := svc.db.Engine.Where("stu_id = ?", account).Or("email = ?", account).Cols("id").Get(&userID)
	if err != nil {
		return 0, err
	}
	if !exist {
		return 0, nil
	}
	return userID, nil
}
