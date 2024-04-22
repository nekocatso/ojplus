package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func getUserIDByContext(ctx *gin.Context) int {
	claims := ctx.Value("claims")
	if claims == nil {
		return 0
	}
	userID := claims.(jwt.MapClaims)["userID"]
	if userID == nil {
		return 0
	}
	return userID.(int)
}
