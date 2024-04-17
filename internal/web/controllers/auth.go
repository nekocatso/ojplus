package controllers

import (
	"Alarm/internal/web/forms"
	"Alarm/internal/web/logs"
	"Alarm/internal/web/models"
	"Alarm/internal/web/services"
	"log"

	"github.com/gin-gonic/gin"
)

type Auth struct {
	svc    *services.Auth
	cfg    map[string]interface{}
	logger *logs.Logger
}

func NewAuth(cfg map[string]interface{}) *Auth {
	svc := services.NewAuth(cfg)
	logger := logs.NewLogger(cfg["db"].(*models.Database))
	return &Auth{svc: svc, cfg: cfg, logger: logger}
}

func (ctrl *Auth) LoginMiddleware(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		response(ctx, 401, nil)
		ctx.Abort()
		return
	}
	claims, err := ctrl.svc.ParseToken(ctrl.cfg["publicKey"], token)
	if err != nil || claims == nil {
		response(ctx, 401, nil)
		ctx.Abort()
		return
	}
	if claims["type"].(string) != "access" {
		response(ctx, 401, nil)
		ctx.Abort()
		return
	}
	if _, ok := claims["userID"]; ok {
		claims["userID"] = int(claims["userID"].(float64))
	}
	user, err := ctrl.svc.GetUserByID(claims["userID"].(int))
	if err != nil {
		response(ctx, 500, nil)
		ctx.Abort()
		return
	}
	if user == nil || !user.IsActive {
		response(ctx, 401, nil)
	}
	ctx.Set("claims", claims)
	ctx.Next()
}

func (ctrl *Auth) AdminMiddleware(ctx *gin.Context) {
	userID := GetUserIDByContext(ctx)
	user, err := ctrl.svc.GetUserByID(userID)
	if err != nil {
		response(ctx, 500, nil)
		ctx.Abort()
		return
	}
	if user.Role < 20 {
		ctx.JSON(404, nil)
		ctx.Abort()
		return
	}
	ctx.Next()
}

func (ctrl *Auth) SuperAdminMiddleware(ctx *gin.Context) {
	userID := GetUserIDByContext(ctx)
	user, err := ctrl.svc.GetUserByID(userID)
	if err != nil {
		response(ctx, 500, nil)
		ctx.Abort()
		return
	}
	if user.Role < 30 {
		ctx.JSON(404, nil)
		ctx.Abort()
		return
	}
	ctx.Next()
}

func (ctrl *Auth) Login(ctx *gin.Context) {
	//校验表单
	form, err := forms.NewLogin(ctx)
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	isValid, errorsMap, err := forms.Verify(form)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if !isValid {
		response(ctx, 400, errorsMap)
		return
	}
	//校验密码
	pass, err := ctrl.svc.VerifyPassword(form.Username, form.Password)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if !pass {
		responseWithMessage(ctx, "用户名或密码错误", 401, nil)
		return
	}
	//获取用户信息
	user, err := ctrl.svc.GetUserByUsername(form.Username)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if !user.IsActive {
		responseWithMessage(ctx, "账户处于禁用状态", 401, nil)
		return
	}
	//生成Token
	accessClaims := map[string]interface{}{
		"userID": user.ID,
		"type":   "access",
	}
	accessToken, err := ctrl.svc.GenerateToken(ctrl.cfg["privateKey"], ctrl.cfg["accessTokenValidity"].(int), accessClaims)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	refreshClaims := map[string]interface{}{
		"userID": user.ID,
		"type":   "refresh",
	}
	refreshToken, err := ctrl.svc.GenerateToken(ctrl.cfg["privateKey"], ctrl.cfg["refreshTokenValidity"].(int), refreshClaims)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	err = ctrl.logger.SaveUserLog(ctx, user, &logs.UserLog{
		Module:  "权限控制",
		Type:    "登录",
		Content: user.Username,
	})
	if err != nil {
		log.Println(err)
	}
	//响应
	data := map[string]interface{}{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
		"userID":       user.ID,
	}
	response(ctx, 200, data)
}

func (ctrl *Auth) Refresh(ctx *gin.Context) {
	form, err := forms.NewRefresh(ctx)
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	isValid, errorsMap, err := forms.Verify(form)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if !isValid {
		response(ctx, 400, errorsMap)
		return
	}
	//解析刷新令牌
	claims, err := ctrl.svc.ParseToken(ctrl.cfg["publicKey"], form.RefreshToken)
	if err != nil || claims == nil {
		response(ctx, 401, nil)
		ctx.Abort()
		return
	}
	if claims["type"].(string) != "refresh" {
		response(ctx, 401, nil)
		ctx.Abort()
		return
	}
	//生成Token
	accessClaims := map[string]interface{}{
		"userID": claims["userID"],
		"type":   "access",
	}
	accessToken, err := ctrl.svc.GenerateToken(ctrl.cfg["privateKey"], ctrl.cfg["accessTokenValidity"].(int), accessClaims)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	refreshClaims := map[string]interface{}{
		"userID": claims["userID"],
		"type":   "refresh",
	}
	refreshToken, err := ctrl.svc.GenerateToken(ctrl.cfg["privateKey"], ctrl.cfg["refreshTokenValidity"].(int), refreshClaims)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	data := map[string]interface{}{
		"accessToken":  accessToken,
		"refreshToken": refreshToken,
	}
	response(ctx, 200, data)
}

func (ctrl *Auth) Test(ctx *gin.Context) {
	response(ctx, 200, ctx.Value("claims"))
}
