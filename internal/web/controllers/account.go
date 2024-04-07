package controllers

import (
	"Alarm/internal/web/forms"
	"Alarm/internal/web/models"
	"Alarm/internal/web/services"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

type Account struct {
	svc *services.Account
	cfg map[string]interface{}
}

func NewAccount(cfg map[string]interface{}) *Account {
	svc := services.NewAccount(cfg)
	return &Account{svc: svc, cfg: cfg}
}

func (ctrl *Account) CreateUser(ctx *gin.Context) {
	form, err := forms.NewUserCreate(ctx)
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	isValid, errorsMap, err := forms.Verify(form)
	if err != nil {
		response(ctx, 500, nil)
		return
	}
	if !isValid {
		response(ctx, 40002, errorsMap)
		return
	}
	user := form.Model
	has, hasMessage, err := ctrl.svc.IsUserExist(user)
	if err != nil {
		response(ctx, 500, nil)
		return
	}
	if has {
		responseWithMessage(ctx, hasMessage, 40901, nil)
		return
	}
	err = ctrl.svc.CreateUser(user)
	if merr, ok := err.(*mysql.MySQLError); ok {
		if merr.Number == 1062 {
			response(ctx, 40901, nil)
			return
		}
	} else if err != nil {
		response(ctx, 500, nil)
	}
	response(ctx, 201, map[string]int{"userID": user.ID})
}

func (ctrl *Account) UpdateUserByID(ctx *gin.Context) {
	// 获取Param参数
	userID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	if userID == 0 {
		response(ctx, 40002, nil)
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
		response(ctx, 500, nil)
		return
	}
	if !isValid {
		response(ctx, 40002, errorsMap)
		return
	}
	//更新数据
	has, err := ctrl.svc.IsUserIDExist(userID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if !has {
		response(ctx, 404, nil)
		return
	}
	err = ctrl.svc.UpdateUserByID(userID, form.Model)
	if err != nil {
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, nil)
}

func (ctrl *Account) FindUsers(ctx *gin.Context) {
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	if pageStr == "" || pageSizeStr == "" {
		response(ctx, 40001, nil)
		return
	}
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		response(ctx, 40002, nil)
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
		response(ctx, 40002, nil)
		return
	}
	if pageSize > 50 {
		response(ctx, 40003, nil)
		return
	}
	users, err := ctrl.svc.AllUser()
	if err != nil {
		response(ctx, 500, nil)
		return
	}
	data := make(map[string]interface{})
	// 对用户列表进行分页处理
	start := (page - 1) * pageSize
	end := start + pageSize
	total := (len(users) + pageSize - 1) / pageSize
	data["total"] = total
	fmt.Println(len(users))
	if start >= len(users) {
		// 响应重定向至最后一页
		redirectURL := fmt.Sprintf("/users?page=%d&pageSize=%d", total, pageSize)
		ctx.Redirect(http.StatusFound, redirectURL)
		return
	}
	if end > len(users) {
		end = len(users)
	}
	data["users"] = users[start:end]
	response(ctx, 200, data)
}

func (ctrl *Account) GetUserByID(ctx *gin.Context) {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		response(ctx, 400, nil)
		return
	}
	var userInfo *models.UserInfo
	userInfo, err = ctrl.svc.GetUserByID(idInt)
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

func (ctrl *Account) GetUsersByAsset(ctx *gin.Context) {
	assetID := ctx.Param("assetID")
	assetIDInt, err := strconv.Atoi(assetID)
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	if assetIDInt <= 0 {
		response(ctx, 40002, nil)
		return
	}
	var users []int
	users, err = ctrl.svc.GetUserIDsByAssetID(assetIDInt)
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
