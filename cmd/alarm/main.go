package main

import (
	"Ojplus/internal/config"
	"Ojplus/internal/utils"
	"Ojplus/internal/web/controllers"
	"Ojplus/internal/web/models"
	"log"

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
	privateKey, err := utils.ReadPrivateKey("./privateKey.pem")
	if err != nil {
		log.Fatal(err)
	}

	publicKey, err := utils.ReadPublicKey("./publicKey.pem")
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
		"accessTokenValidity":  globalConfig.Gin.Token.AccessValidity,
		"refreshTokenValidity": globalConfig.Gin.Token.RefreshValidity,
	}
	AuthCtrl := controllers.NewAuth(authConfig)

	// --Router Init
	group := engine.Group("/api/v1")
	{
		group.POST("/user", AuthCtrl.CreateUser)
		group.PATCH("/users/:id", AuthCtrl.LoginMiddleware, AuthCtrl.UpdateUser)
		group.DELETE("/users/:id", AuthCtrl.LoginMiddleware, AuthCtrl.DeleteUser)

		// group.GET("/auth/test/login", AuthCtrl.LoginMiddleware, AuthCtrl.Test)
		group.POST("/token", AuthCtrl.Login)
		group.PATCH("/token", AuthCtrl.Refresh)
	}
	engine.Run(globalConfig.Gin.Port)
}
