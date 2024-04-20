package controllers

import (
	"Ojplus/internal/web/forms"
	"Ojplus/internal/web/services"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

type Auth struct {
	svc *services.Auth
	cfg map[string]any
}

func NewAuth(cfg map[string]any) *Auth {
	svc := services.NewAuth(cfg)
	return &Auth{svc: svc, cfg: cfg}
}

// Auth
func (ctrl *Auth) LoginMiddleware(ctx *gin.Context) {
	token := ctx.GetHeader("Authorization")
	if token == "" {
		response(ctx, 401, nil)
		ctx.Abort()
		return
	}
	claims, err := ctrl.svc.ParseToken(token)
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
	userID := int(claims["userID"].(float64))
	if _, ok := claims["userID"]; ok {
		claims["userID"] = userID
	}
	pass, err := ctrl.svc.CheckPermisson(userID, 0)
	if err != nil {
		response(ctx, 500, nil)
		ctx.Abort()
		return
	}
	if !pass {
		response(ctx, 401, nil)
		ctx.Abort()
		return
	}
	ctx.Set("claims", claims)
	ctx.Next()
}

func (ctrl *Auth) AdminMiddleware(ctx *gin.Context) {
	userID := getUserIDByContext(ctx)
	pass, err := ctrl.svc.CheckPermisson(userID, 20)
	if err != nil {
		response(ctx, 500, nil)
		ctx.Abort()
		return
	}
	if !pass {
		ctx.JSON(404, nil)
		ctx.Abort()
		return
	}
	ctx.Next()
}

func (ctrl *Auth) SuperAdminMiddleware(ctx *gin.Context) {
	userID := getUserIDByContext(ctx)
	pass, err := ctrl.svc.CheckPermisson(userID, 30)
	if err != nil {
		response(ctx, 500, nil)
		ctx.Abort()
		return
	}
	if !pass {
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
	userID, err := ctrl.svc.GetUserIDByAccount(form.Account)
	if err != nil {
		response(ctx, 500, nil)
		return
	}
	if userID == 0 {
		response(ctx, 404, nil)
		return
	}
	pass, err := ctrl.svc.VerifyPassword(userID, form.Password)
	if err != nil {
		response(ctx, 500, nil)
		return
	}
	if !pass {
		response(ctx, 401, nil)
		return
	}
	//生成Token
	data, err := ctrl.svc.GenerateToken(userID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
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
	claims, err := ctrl.svc.ParseToken(form.Token)
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
	userID := claims["userID"].(int)
	data, err := ctrl.svc.GenerateToken(userID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, data)
}

// User
func (ctrl *Auth) CreateUser(ctx *gin.Context) {
	form, err := forms.NewUserCreate(ctx)
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
		response(ctx, 40002, errorsMap)
		return
	}
	msg, err := ctrl.svc.GetUserExistInfo(*form.Username, *form.Email)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if msg != "" {
		responseWithMessage(ctx, msg, 40901, nil)
		return
	}
	userID, err := ctrl.svc.CreateUser(form)
	if merr, ok := err.(*mysql.MySQLError); ok {
		if merr.Number == 1062 {
			response(ctx, 40901, nil)
			return
		}
	} else if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 201, map[string]int{"userID": userID})
}

func (ctrl *Auth) UpdateUser(ctx *gin.Context) {
	// 用户校验
	userID := getUserIDByContext(ctx)
	// 表单校验
	form, err := forms.NewUserUpdate(ctx)
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
		response(ctx, 40002, errorsMap)
		return
	}
	// 修改密码的处理
	if form.Password != nil && form.OldPassword != nil {
		pass, err := ctrl.svc.VerifyPassword(userID, *form.OldPassword)
		if err != nil {
			log.Println(err)
			response(ctx, 500, nil)
			return
		}
		if !pass {
			response(ctx, 40101, nil)
			return
		}
	}
	//更新数据
	err = ctrl.svc.UpdateUser(userID, form)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, nil)
}

func (ctrl *Auth) DeleteUser(ctx *gin.Context) {
	form, err := forms.NewUserCreate(ctx)
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
		response(ctx, 40002, errorsMap)
		return
	}
	// 数据校验
	userID := getUserIDByContext(ctx)
	err = ctrl.svc.DeleteUser(userID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, nil)
}

// func (ctrl *Auth) GetUsers(ctx *gin.Context) {
// 	pageStr := ctx.Query("page")
// 	pageSizeStr := ctx.Query("pageSize")
// 	var page, pageSize int
// 	if pageStr != "" && pageSizeStr != "" {
// 		var err error
// 		page, err = strconv.Atoi(pageStr)
// 		if err != nil || page <= 0 || pageSize > 100 {
// 			response(ctx, 40002, nil)
// 			return
// 		}
// 		pageSize, err = strconv.Atoi(pageSizeStr)
// 		if err != nil || pageSize <= 0 {
// 			response(ctx, 40002, nil)
// 			return
// 		}
// 	} else {
// 		page = 1
// 		pageSize = 20
// 	}
// 	users, err := ctrl.svc.GetUsers(map[string]any{})
// 	if err != nil {
// 		response(ctx, 500, nil)
// 		return
// 	}
// 	// 对用户列表进行分页处理
// 	data := make(map[string]any)
// 	start := (page - 1) * pageSize
// 	end := start + pageSize
// 	pages := (len(users) + pageSize - 1) / pageSize
// 	if pages == 0 {
// 		pages = 1
// 	}
// 	data["pages"] = pages
// 	data["total"] = len(users)
// 	if start >= len(users) {
// 		// 响应最后一页
// 		start = (pages - 1) * pageSize
// 		end = len(users)
// 	}
// 	if end > len(users) {
// 		end = len(users)
// 	}
// 	data["users"] = users[start:end]
// 	response(ctx, 200, data)
// }
