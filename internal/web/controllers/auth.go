package controllers

import (
	"Alarm/internal/web/models"
	"Alarm/internal/web/services"

	"github.com/gin-gonic/gin"
)

type Auth struct {
	svc *services.Auth
}

func NewAuth(db *models.Database) *Auth {
	svc := services.NewAuth(db)
	return &Auth{svc: svc}
}

func (ctrl *Auth) Login(ctx *gin.Context) {
	ctrl.svc.CreateToken()
}
func (ctrl *Auth) Logout(ctx *gin.Context) {

}
func (ctrl *Auth) Register(ctx *gin.Context) {

}
