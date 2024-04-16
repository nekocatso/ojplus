package controllers

import (
	"Alarm/internal/web/forms"
	"Alarm/internal/web/logs"
	"Alarm/internal/web/models"
	"Alarm/internal/web/services"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

type Account struct {
	svc    *services.Account
	cfg    map[string]interface{}
	logger *logs.Logger
}

func NewAccount(cfg map[string]interface{}) *Account {
	svc := services.NewAccount(cfg)
	logger := logs.NewLogger(cfg["db"].(*models.Database))
	return &Account{svc: svc, cfg: cfg, logger: logger}
}

func (ctrl *Account) CreateUser(ctx *gin.Context) {
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
	user := form.Model
	has, hasMessage, err := ctrl.svc.GetUserExistInfo(user)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if has {
		responseWithMessage(ctx, hasMessage, 40901, nil)
		return
	}
	user.IsActive = true
	err = ctrl.svc.CreateUser(user)
	if merr, ok := err.(*mysql.MySQLError); ok {
		if merr.Number == 1062 {
			response(ctx, 40901, nil)
			return
		}
	} else if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
	}
	response(ctx, 201, map[string]int{"userID": user.ID})
	err = ctrl.logger.SaveUserLog(ctx, user.ID, &logs.UserLog{
		Module:  "账号管理",
		Type:    "新增",
		Content: user.Username,
	})
	if err != nil {
		log.Println(err)
	}
}

func (ctrl *Account) UpdateUser(ctx *gin.Context) {
	// 参数校验
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	if userID <= 0 {
		response(ctx, 40002, nil)
		return
	}
	// 权限校验
	user, err := ctrl.svc.GetUserByID(userID)
	if err != nil {
		log.Println(err)
		log.Println("hi1")
		response(ctx, 500, nil)
		return
	}
	if user == nil {
		log.Println(userID)
		response(ctx, 404, nil)
		return
	}
	loginerID := GetUserIDByContext(ctx)
	loginer, err := ctrl.svc.GetUserByID(loginerID)
	if err != nil {
		log.Println(err)
		log.Println("hi2")
		response(ctx, 500, nil)
		return
	}
	if userID != loginerID && loginer.Role < 30 {
		log.Println(userID, loginerID, loginer.Role)
		response(ctx, 404, nil)
		return
	}
	// 表单校验
	form, err := forms.NewUserUpdate(ctx)
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	isValid, errorsMap, err := forms.Verify(form)
	if err != nil {
		log.Println(err)
		log.Println("hi3")
		response(ctx, 500, nil)
		return
	}
	if !isValid {
		response(ctx, 40002, errorsMap)
		return
	}
	// 重置密码的处理
	if form.IsResetPwd && user.Role >= 30 {
		err := ctrl.svc.RestPassword(userID)
		if err != nil {
			log.Println(err)
			log.Println(userID, loginerID)
			log.Println("hi4")
			response(ctx, 500, nil)
			return
		}
	}
	// 修改密码的处理
	if form.Password != "" && form.OldPassword != "" {
		pass, err := ctrl.svc.VerifyPassword(user.Username, form.OldPassword)
		if err != nil {
			log.Println(err)
			log.Println("hi5")
			response(ctx, 500, nil)
			return
		}
		if !pass {
			response(ctx, 40101, nil)
			return
		}
		form.UpdateMap["password"] = form.Password
	}
	//更新数据
	err = ctrl.svc.UpdateUserByID(userID, form.UpdateMap)
	if err != nil {
		log.Println(err)
		log.Println("hi6")
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, nil)
	err = ctrl.logger.SaveUserLog(ctx, loginerID, &logs.UserLog{
		Module:  "账号管理",
		Type:    "编辑",
		Content: user.Username,
	})
	if err != nil {
		log.Println(err)
	}
}

func (ctrl *Account) GetUsers(ctx *gin.Context) {
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	var page, pageSize int
	if pageStr != "" && pageSizeStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page <= 0 || pageSize > 100 {
			response(ctx, 40002, nil)
			return
		}
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize <= 0 {
			response(ctx, 40002, nil)
			return
		}
	} else {
		page = 1
		pageSize = 10
	}
	users, err := ctrl.svc.FindUsers(map[string]interface{}{})
	if err != nil {
		response(ctx, 500, nil)
		return
	}
	// 对用户列表进行分页处理
	data := make(map[string]interface{})
	start := (page - 1) * pageSize
	end := start + pageSize
	pages := (len(users) + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}
	data["pages"] = pages
	data["total"] = len(users)
	if start >= len(users) {
		// 响应最后一页
		start = (pages - 1) * pageSize
		end = len(users)
	}
	if end > len(users) {
		end = len(users)
	}
	data["users"] = users[start:end]
	response(ctx, 200, data)
}

func (ctrl *Account) SelectUsers(ctx *gin.Context) {
	//校验数据
	form, err := forms.NewUserSelect(ctx)
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
	page := form.Page
	pageSize := form.PageSize

	users, err := ctrl.svc.FindUsers(form.Conditions)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	// 对用户列表进行分页处理
	data := make(map[string]interface{})
	start := (page - 1) * pageSize
	end := start + pageSize
	pages := (len(users) + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}
	data["pages"] = pages
	data["total"] = len(users)
	if start >= len(users) {
		// 响应最后一页
		start = (pages - 1) * pageSize
		end = len(users)
	}
	if end > len(users) {
		end = len(users)
	}
	data["users"] = users[start:end]
	response(ctx, 200, data)
}

func (ctrl *Account) GetUserByID(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	if id <= 0 {
		response(ctx, 40002, nil)
		return
	}
	var userInfo *models.UserInfo
	userInfo, err = ctrl.svc.GetUserByID(id)
	if err != nil {
		response(ctx, 500, nil)
		log.Println(err)
		return
	}
	if userInfo == nil || userInfo.ID == 0 {
		response(ctx, 404, nil)
		return
	}
	response(ctx, 200, userInfo)
}

func (ctrl *Account) GetUserIDsByAssetID(ctx *gin.Context) {
	assetID, err := strconv.Atoi(ctx.Param("assetID"))
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	if assetID <= 0 {
		response(ctx, 40002, nil)
		return
	}
	var users []int
	users, err = ctrl.svc.GetUserIDsByAssetID(assetID)
	if err != nil {
		response(ctx, 500, nil)
		log.Println(err)
		return
	}
	if users == nil {
		response(ctx, 404, nil)
		return
	}
	response(ctx, 200, users)
}
