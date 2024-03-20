package main

import (
	"Alarm/internal/config"
	"Alarm/internal/web/controllers"
	"Alarm/internal/web/models"
	"Alarm/internal/web/routers"
	"log"
)

func main() {
	// Config Init
	config, err := config.NewConfig(".", "config")
	if err != nil {
		log.Fatal(err)
	}

	// Mysql Init
	db, err := models.NewDatabase(config.Mysql)
	if err != nil {
		log.Fatal(err)
	}

	// Gin routers Init
	router := routers.NewRouter()
	router.AccountCtrl(controllers.NewAccountController(db))
	router.AuthCtrl(controllers.NewAuthController(db))
	router.Run(config.Gin.Port)
}
