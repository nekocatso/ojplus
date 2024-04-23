package main

import (
	"Ojplus/internal/config"
	"Ojplus/internal/utils"
	"Ojplus/internal/web/controllers"
	"Ojplus/internal/web/models"
	"flag"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	configFile := flag.Bool("p", false, "Use config-main.toml")
	flag.Parse()

	var configFileName string
	if *configFile {
		configFileName = "config-main"
	} else {
		configFileName = "config"
	}

	// Config Init
	globalConfig, err := config.NewConfig(".", configFileName)
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
	// Email Init
	templatePaths := map[string]string{
		"verificationTemplate": "./assets/email-templates/verification.html",
	}
	globalConfig.Gin.Email.Init(templatePaths)
	// Auth Init
	privateKey, err := utils.ReadPrivateKey("./assets/rsa-keys/privateKey.pem")
	if err != nil {
		log.Fatal(err)
	}

	publicKey, err := utils.ReadPublicKey("./assets/rsa-keys/publicKey.pem")
	if err != nil {
		log.Fatal(err)
	}

	// Gin Init
	engine := gin.Default()

	// --Controller Init
	// ctrlConfig := map[string]any{
	// 	"db":     db,
	// 	"cache":  cache,
	// 	"global": globalConfig,
	// }
	authConfig := map[string]any{
		"db":                   db,
		"cache":                cache,
		"privateKey":           privateKey,
		"publicKey":            publicKey,
		"global":               globalConfig,
		"accessTokenValidity":  globalConfig.Gin.Token.AccessExp,
		"refreshTokenValidity": globalConfig.Gin.Token.RefreshExp,
	}
	AuthCtrl := controllers.NewAuth(authConfig)

	// --Router Init
	group := engine.Group("/api/v1")
	{
		group.POST("/user", AuthCtrl.CreateUser)
		group.PATCH("/users/:userID", AuthCtrl.LoginMiddleware, AuthCtrl.UpdateUser)
		group.DELETE("/users/:userID", AuthCtrl.LoginMiddleware, AuthCtrl.DeleteUser)

		group.GET("/auth/test/login", AuthCtrl.LoginMiddleware, AuthCtrl.Test)
		group.POST("/token", AuthCtrl.TokenCreate)
		group.PATCH("/token", AuthCtrl.TokenRefresh)

		group.POST("/verification", AuthCtrl.CreateVerification)

	}
	engine.Run(globalConfig.Gin.Port)
}
