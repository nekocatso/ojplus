package controllers

import (
	"Alarm/example/services"

	"github.com/gin-gonic/gin"
)

type Auth struct {
	svc *services.Auth
}

func NewController(svc *services.Auth) *Auth {
	return &Auth{svc: svc}
}

func (ctrl *Auth) SignUp(ctx *gin.Context) {
	ctrl.svc.CreateToken()
}
