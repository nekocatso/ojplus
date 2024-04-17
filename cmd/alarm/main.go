package main

import (
	"Alarm/internal/config"
	"Alarm/internal/pkg/listenerpool"
	"Alarm/internal/pkg/mail"
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

	// mailBox := []mail.MailBox{}
	// for _, mailConfig := range globalConfig.Listener.Mails {
	// 	mailBox = append(mailBox, mail.MailBox{
	// 		Name:     mailConfig.Name,
	// 		Password: mailConfig.Password,
	// 		Host:     mailConfig.Host,
	// 		Port:     mailConfig.Port,
	// 	})
	// }
	mailBox := []mail.MailBox{{
		Name:     "yangquanmailtest@163.com",
		Password: "APQJNHKHMXPGRFVO",
		Host:     "smtp.163.com",
		Port:     25,
	}}
	// Mail Init
	mail, err := mail.NewMailPool(mailBox)
	if err != nil {
		log.Fatal(err)
	}
	// ListeningPool Init
	listener, err := listenerpool.NewListenerPool(db, cache, mail, "amqp://user:mkjsix7@172.16.0.15:5672/")
	if err != nil {
		log.Fatal(err)
	}
	// --Controller Init
	ctrlConfig := map[string]interface{}{
		"db":       db,
		"cache":    cache,
		"listener": listener,
		"global":   globalConfig,
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
	LogCtrl := controllers.NewLog(ctrlConfig)

	// --Router Init
	group := engine.Group("/api")
	{
		group.POST("/register", AccountCtrl.CreateUser, AuthCtrl.LoginMiddleware, AuthCtrl.SuperAdminMiddleware)
		group.POST("/users/query", AuthCtrl.LoginMiddleware, AccountCtrl.SelectUsers)
		group.GET("/users", AuthCtrl.LoginMiddleware, AccountCtrl.GetUsers)
		group.GET("/user/:id", AuthCtrl.LoginMiddleware, AccountCtrl.GetUserByID)
		group.PATCH("/user/:id", AuthCtrl.LoginMiddleware, AuthCtrl.AdminMiddleware, AccountCtrl.UpdateUser)
		group.DELETE("/user/:id", AuthCtrl.LoginMiddleware, AuthCtrl.AdminMiddleware, AccountCtrl.DeleteUser)

		group.GET("/authtest", AuthCtrl.LoginMiddleware, AuthCtrl.Test)
		group.POST("/login", AuthCtrl.Login)
		group.POST("/token", AuthCtrl.Refresh)

		group.POST("/asset", AuthCtrl.LoginMiddleware, AuthCtrl.AdminMiddleware, AssetCtrl.CreateAsset)
		group.POST("/assets/query", AuthCtrl.LoginMiddleware, AssetCtrl.SelectAssets)
		group.GET("/assets", AuthCtrl.LoginMiddleware, AssetCtrl.GetAssets)
		group.GET("/assets/info", AuthCtrl.LoginMiddleware, AssetCtrl.GetAssetsInfo)
		group.GET("/asset/:id", AuthCtrl.LoginMiddleware, AssetCtrl.GetAssetByID)
		group.GET("/assets/id", AuthCtrl.LoginMiddleware, AssetCtrl.GetAssetIDs)
		group.PATCH("/asset/:id", AuthCtrl.LoginMiddleware, AuthCtrl.AdminMiddleware, AssetCtrl.UpdateAssetByID)
		group.DELETE("/asset/:id", AuthCtrl.LoginMiddleware, AuthCtrl.AdminMiddleware, AssetCtrl.DeleteAsset)
		group.GET("/assets/:assetID/:target", AuthCtrl.LoginMiddleware, func(ctx *gin.Context) {
			if ctx.Param("target") == "users" {
				AccountCtrl.GetUsersByAssetID(ctx)
			} else if ctx.Param("target") == "rules" {
				RuleCtrl.GetRulesByAssetID(ctx)
			} else {
				ctx.JSON(404, nil)
			}
		})

		group.GET("/rules", AuthCtrl.LoginMiddleware, RuleCtrl.GetRules)
		group.GET("/rule/:id", AuthCtrl.LoginMiddleware, RuleCtrl.GetRuleByID)
		group.POST("/rule", AuthCtrl.LoginMiddleware, AuthCtrl.AdminMiddleware, RuleCtrl.CreateRule)
		group.POST("/rules/query", AuthCtrl.LoginMiddleware, RuleCtrl.SelectRules)
		group.PATCH("/rule/:id", AuthCtrl.LoginMiddleware, AuthCtrl.AdminMiddleware, RuleCtrl.UpdateRuleByID)
		group.DELETE("/rule/:id", AuthCtrl.LoginMiddleware, AuthCtrl.AdminMiddleware, RuleCtrl.DeleteRuleByID)
		group.GET("/rules/:ruleID/:target", AuthCtrl.LoginMiddleware, func(ctx *gin.Context) {
			if ctx.Param("target") == "assets" {
				AssetCtrl.GetAssetsByRuleID(ctx)
			} else {
				ctx.JSON(404, nil)
			}
		})

		group.POST("/alarm", AuthCtrl.LoginMiddleware, AuthCtrl.AdminMiddleware, AlarmCtrl.CreateAlarm)
		group.GET("/alarms", AuthCtrl.LoginMiddleware, AlarmCtrl.GetAlarms)
		group.GET("/alarm/:id", AuthCtrl.LoginMiddleware, AlarmCtrl.GetAlarmByID)
		group.POST("/alarms/query", AuthCtrl.LoginMiddleware, AlarmCtrl.SelectAlarms)
		group.DELETE("/alarm/:id", AuthCtrl.LoginMiddleware, AuthCtrl.AdminMiddleware, AlarmCtrl.DeleteAlarmByID)
		group.PATCH("/alarm/:id", AuthCtrl.LoginMiddleware, AlarmCtrl.UpdateAlarmByID)
		group.GET("/alarms/:alarmID/:target", AuthCtrl.LoginMiddleware, func(ctx *gin.Context) {
			if ctx.Param("target") == "rules" {
				RuleCtrl.GetRulesByAlarmID(ctx)
			} else {
				ctx.JSON(404, nil)
			}
		})

		group.GET("/log/alarms", AuthCtrl.LoginMiddleware, LogCtrl.GetAlarmLogs)
		group.GET("/log/alarm/:id", AuthCtrl.LoginMiddleware, LogCtrl.GetAlarmLogByID)
		group.POST("/log/alarm/info", AuthCtrl.LoginMiddleware, LogCtrl.GetAlarmLogInfo)
		group.POST("/log/alarms/query", AuthCtrl.LoginMiddleware, LogCtrl.SelectAlarmLogs)
		group.GET("/log/users", AuthCtrl.LoginMiddleware, LogCtrl.GetUserLogs)
		group.POST("/log/user", AuthCtrl.LoginMiddleware, LogCtrl.CreateUserLog)
		group.POST("/log/users/query", AuthCtrl.LoginMiddleware, LogCtrl.SelectUserLogs)
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
