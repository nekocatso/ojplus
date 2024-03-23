package main

import (
	"Alarm/internal/config"
	"Alarm/internal/web/controllers"
	"Alarm/internal/web/models"
	"Alarm/internal/web/routers"
	"log"

	_ "github.com/go-sql-driver/mysql"
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
	cache, err := models.NewCache(config.Redis)
	if err != nil {
		log.Fatal(err)
	}

	// Gin routers Init
	router := routers.NewRouter()
	router.AccountCtrl(controllers.NewAccountController(db, cache))
	router.AuthCtrl(controllers.NewAuthController(db, cache))
	router.Run(config.Port)
}
