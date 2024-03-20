package controllers

import (
	"Alarm/internal/web/models"
	"Alarm/internal/web/services"

	"github.com/gin-gonic/gin"
)

type AccountController struct {
	svc *services.AccountService
}

func NewAccountController(db *models.Database) *AccountController {
	svc := services.NewAccount(db)
	return &AccountController{svc: svc}
}

func (ctrl *AccountController) Register(ctx *gin.Context) {
	ctrl.svc.CreateUser()
}
