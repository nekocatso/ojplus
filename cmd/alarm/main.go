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

	// Database Init
	db, err := models.NewDatabase(config.Mysql)
	if err != nil {
		log.Fatal(err)
	}
	router := routers.NewRouter()

	// Auth Module
	authController := controllers.NewAuth(db)
	router.AuthInit(authController)
	router.Run()
}
