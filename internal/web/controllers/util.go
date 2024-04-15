package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func GetUserIDByContext(ctx *gin.Context) int {
	claims := ctx.Value("claims").(jwt.MapClaims)
	return claims["userID"].(int)
}
