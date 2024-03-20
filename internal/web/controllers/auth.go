package controllers

import (
	"Alarm/internal/web/models"
	"Alarm/internal/web/services"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	svc *services.AuthService
}

func NewAuthController(db *models.Database) *AuthController {
	svc := services.NewAuth(db)
	return &AuthController{svc: svc}
}

func (ctrl *AuthController) Login(ctx *gin.Context) {
	ctrl.svc.CreateToken()
}
func (ctrl *AuthController) Logout(ctx *gin.Context) {

}
