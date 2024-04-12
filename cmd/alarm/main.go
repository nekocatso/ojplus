package main

import (
	"Alarm/internal/config"
	"Alarm/internal/web/controllers"
	"Alarm/internal/web/models"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// Config Init
	globalConfig, err := config.NewConfig(".", "config")
	if err != nil {
		log.Fatal(err)
	}

	// Mysql Init
	db, err := models.NewDatabase(globalConfig.Gin.Mysql)
	if err != nil {
		log.Fatal(err)
	}

	//Redis Init
	cache, err := models.NewCache(globalConfig.Gin.Redis)
	if err != nil {
		log.Fatal(err)
	}

	// Auth Init
	privateKey, err := readPrivateKeyFromFile("./privateKey.pem")
	if err != nil {
		log.Fatal(err)
	}

	publicKey, err := readPublicKeyFromFile("./publicKey.pem")
	if err != nil {
		log.Fatal(err)
	}

	// Gin Init
	engine := gin.Default()

	// --Controller Init
	ctrlConfig := map[string]interface{}{
		"db":    db,
		"cache": cache,
	}
	authConfig := map[string]interface{}{
		"db":                   db,
		"cache":                cache,
		"privateKey":           privateKey,
		"publicKey":            publicKey,
		"accessTokenValidity":  globalConfig.Gin.Token.AccessValidity,
		"refreshTokenValidity": globalConfig.Gin.Token.RefreshValidity,
	}
	AccountCtrl := controllers.NewAccount(ctrlConfig)
	AuthCtrl := controllers.NewAuth(authConfig)
	AssetCtrl := controllers.NewAsset(ctrlConfig)
	RuleCtrl := controllers.NewRule(ctrlConfig)
	AlarmCtrl := controllers.NewAlarm(ctrlConfig)

	// --Router Init
	group := engine.Group("/api")
	{
		group.POST("/register", AccountCtrl.CreateUser)
		group.POST("/users/query", AccountCtrl.SelectUsers)
		group.GET("/users", AuthCtrl.LoginMiddleware, AccountCtrl.GetUsers)
		group.GET("/user/:id", AuthCtrl.LoginMiddleware, AccountCtrl.GetUserByID)
		group.PATCH("/user/:id", AuthCtrl.LoginMiddleware, AccountCtrl.UpdateUser)

		group.GET("/authtest", AuthCtrl.LoginMiddleware, AuthCtrl.Test)
		group.POST("/login", AuthCtrl.Login)
		group.POST("/token", AuthCtrl.Refresh)

		group.GET("/assets", AuthCtrl.LoginMiddleware, AssetCtrl.GetAssets)
		group.GET("/asset/:id", AuthCtrl.LoginMiddleware, AssetCtrl.GetAssetByID)
		group.GET("/assets/id", AuthCtrl.LoginMiddleware, AssetCtrl.GetAssetIDs)
		group.POST("/asset", AuthCtrl.LoginMiddleware, AssetCtrl.CreateAsset)
		group.POST("/assets/query", AuthCtrl.LoginMiddleware, AssetCtrl.SelectAssets)
		// group.GET("/assets/:assetID/:target", AuthCtrl.LoginMiddleware, AccountCtrl.GetUsersByAsset)
		group.GET("/assets/:assetID/:target", AuthCtrl.LoginMiddleware, func(ctx *gin.Context) {
			if ctx.Param("target") == "users" {
				AccountCtrl.GetUserIDsByAssetID(ctx)
			} else {
				RuleCtrl.GetRuleIDsByAssetID(ctx)
			}
		})

		group.GET("/rules", AuthCtrl.LoginMiddleware, RuleCtrl.GetRules)
		group.POST("/rule", AuthCtrl.LoginMiddleware, RuleCtrl.CreateRule)
		group.POST("/rules/query", AuthCtrl.LoginMiddleware, RuleCtrl.SelectRules)

		group.POST("/alarm", AuthCtrl.LoginMiddleware, AlarmCtrl.CreateAlarm)
	}
	engine.Run(globalConfig.Gin.Port)
}

func readPrivateKeyFromFile(filepath string) (interface{}, error) {
	keyFile, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyFile)
	if block == nil {
		return nil, fmt.Errorf("decode private key error")
	}

	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func readPublicKeyFromFile(filepath string) (interface{}, error) {
	keyFile, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyFile)
	if block == nil {
		return nil, fmt.Errorf("decode public key error")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return key, nil
}
