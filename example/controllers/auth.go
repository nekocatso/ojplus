package controllers

import (
	"Alarm/example/services"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	svc *services.AuthService
}

func NewController(svc *services.AuthService) *AuthController {
	return &AuthController{svc: svc}
}

func (ctrl *AuthController) SignUp(ctx *gin.Context) {
	ctrl.svc.CreateToken()
}
