package services

import (
	"Alarm/internal/web/models"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Auth struct {
	db    *models.Database
	cache *models.Cache
	cfg   map[string]interface{}
}

func NewAuth(cfg map[string]interface{}) *Auth {
	return &Auth{
		db:    cfg["db"].(*models.Database),
		cache: cfg["cache"].(*models.Cache),
		cfg:   cfg,
	}
}

// RefreshToken 用于刷新token，可以创建新的token或者延长现有token的过期时间。
//
// 如果 tokenStr 为空字符串，则使用私钥创建新的token。
func (svc *Auth) GenerateToken(privateKey interface{}, validSeconds int, data map[string]interface{}) (string, error) {
	claims := jwt.MapClaims{
		"exp": time.Now().Add(time.Second * time.Duration(validSeconds)).Unix(),
	}

	for key, value := range data {
		claims[key] = value
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	tokenStr, err := token.SignedString(privateKey)
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (svc *Auth) ParseToken(publicKey interface{}, tokenStr string) (jwt.MapClaims, error) {
	// 解析令牌
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法是否为RS256
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("invalid signing method: %v", token.Header["alg"])
		}

		return publicKey, nil
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

	return claims, nil
}
func (svc *Auth) VerifyPassword(username string, password string) (bool, error) {
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

func (svc *Auth) GetUserByUsername(username string) (*models.User, error) {
	user := &models.User{Username: username}
	_, err := svc.db.Engine.Get(user)
	if err != nil {
		return nil, err
	}
	userInfo := user
	return userInfo, nil
}

func (svc *Auth) GetUserByID(userID int) (*models.User, error) {
	return GetUserByID(svc.db.Engine, userID)
}
