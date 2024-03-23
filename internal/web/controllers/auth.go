package controllers

import (
	"Alarm/internal/web/models"
	"Alarm/internal/web/services"

	"github.com/gin-gonic/gin"
)

type Auth struct {
	svc *services.Auth
}

func NewAuthController(db *models.Database, cache *models.Cache) *Auth {
	svc := services.NewAuth(db, cache)
	return &Auth{svc: svc}
}

func (ctrl *Auth) Login(ctx *gin.Context) {
	// ctrl.svc.RefreshToken()
}
func (ctrl *Auth) Logout(ctx *gin.Context) {

}
