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

	// --Router Init
	group := engine.Group("")
	{
		group.POST("/register", AccountCtrl.CreateUser)
		group.GET("/users", AuthCtrl.LoginMiddleware, AccountCtrl.FindUsers)
		group.GET("/users/:id", AuthCtrl.LoginMiddleware, AccountCtrl.GetUserByID)
		group.GET("/assets/:assetID/users", AuthCtrl.LoginMiddleware, AccountCtrl.GetUsersByAsset)
		group.PATCH("/users/:id", AuthCtrl.LoginMiddleware, AccountCtrl.UpdateUser)

		group.GET("/authtest", AuthCtrl.LoginMiddleware, AuthCtrl.Test)
		group.POST("/login", AuthCtrl.Login)
		group.POST("/token", AuthCtrl.Refresh)

		group.POST("/asset", AuthCtrl.LoginMiddleware, AssetCtrl.CreateAsset)
		group.POST("/assets/query", AuthCtrl.LoginMiddleware, AssetCtrl.SelectAsset)
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
