package controllers

import (
	"Alarm/internal/web/forms"
	"Alarm/internal/web/services"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

type Alarm struct {
	svc *services.Alarm
	cfg map[string]interface{}
}

func NewAlarm(cfg map[string]interface{}) *Alarm {
	svc := services.NewAlarm(cfg)
	return &Alarm{svc: svc, cfg: cfg}
}

func (ctrl *Alarm) CreateAlarm(ctx *gin.Context) {
	form, err := forms.NewAlarmCreate(ctx)
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
	alarm := form.Model
	err = ctrl.svc.CreateAlarm(alarm)
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
	has, err := ctrl.svc.SetAlarm(alarm)
	if err != nil || !has {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 201, map[string]int{"ararmID": alarm.ID})
}

func (ctrl *Alarm) GetAlarms(ctx *gin.Context) {
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
	userID := GetUserIDByContext(ctx)
	alarms, err := ctrl.svc.FindAlarms(userID, map[string]interface{}{})
	if err != nil {
		response(ctx, 500, nil)
		return
	}
	// 分页处理
	data := make(map[string]interface{})
	start := (page - 1) * pageSize
	end := start + pageSize
	pages := (len(alarms) + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}
	data["pages"] = pages
	data["total"] = len(alarms)
	if start >= len(alarms) {
		// 响应最后一页
		start = (pages - 1) * pageSize
		end = len(alarms)
	}
	if end > len(alarms) {
		end = len(alarms)
	}
	data["alarms"] = alarms[start:end]
	response(ctx, 200, data)
}

func (ctrl *Alarm) SelectAlarms(ctx *gin.Context) {
	//校验数据
	form, err := forms.NewAlarmSelect(ctx)
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
	userID := GetUserIDByContext(ctx)
	users, err := ctrl.svc.FindAlarms(userID, form.Conditions)
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
