package controllers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func getUserIDByContext(ctx *gin.Context) int {
	claims := ctx.Value("claims")
	if claims == nil {
		return 0
	}

	switch claims.(type) {
	case jwt.MapClaims:
		userID := claims.(jwt.MapClaims)["userID"]
		if userID == nil {
			return 0
		}
		return userID.(int)
	case map[string]interface{}:
		userID := claims.(map[string]interface{})["userID"]
		if userID == nil {
			return 0
		}
		return userID.(int)
	default:
		return 0
	}
}
func getPageInfoByContext(ctx *gin.Context) (int, int) {
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	var page, pageSize int
	if pageStr != "" && pageSizeStr != "" {
		page, _ = strconv.Atoi(pageStr)
		pageSize, _ = strconv.Atoi(pageSizeStr)
		if page < 1 {
			page = 1
		}
		if pageSize < 5 {
			pageSize = 5
		} else if pageSize > 200 {
			pageSize = 200
		}
	} else {
		page = 1
		pageSize = 20
	}
	return page, pageSize
}
