package controllers

import (
	"Alarm/internal/web/forms"
	"Alarm/internal/web/services"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Log struct {
	svc *services.Log
	cfg map[string]interface{}
}

func NewLog(cfg map[string]interface{}) *Log {
	svc := services.NewLog(cfg)
	return &Log{svc: svc, cfg: cfg}
}

func (ctrl *Log) GetAlarmLogs(ctx *gin.Context) {
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
	alarmLogs, err := ctrl.svc.FindALarmLogs(userID, map[string]interface{}{})
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	// 对用户列表进行分页处理
	data := make(map[string]interface{})
	start := (page - 1) * pageSize
	end := start + pageSize
	pages := (len(alarmLogs) + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}
	data["pages"] = pages
	data["total"] = len(alarmLogs)
	if start >= len(alarmLogs) {
		// 响应最后一页
		start = (pages - 1) * pageSize
		end = len(alarmLogs)
	}
	if end > len(alarmLogs) {
		end = len(alarmLogs)
	}
	data["logs"] = alarmLogs[start:end]
	response(ctx, 200, data)
}

func (ctrl *Log) SelectAlarmLogs(ctx *gin.Context) {
	form, err := forms.NewAlarmLogSelect(ctx)
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
	alarmLogs, err := ctrl.svc.FindALarmLogs(userID, form.Conditions)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	// 对用户列表进行分页处理
	data := make(map[string]interface{})
	start := (page - 1) * pageSize
	end := start + pageSize
	pages := (len(alarmLogs) + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}
	data["pages"] = pages
	data["total"] = len(alarmLogs)
	if start >= len(alarmLogs) {
		// 响应最后一页
		start = (pages - 1) * pageSize
		end = len(alarmLogs)
	}
	if end > len(alarmLogs) {
		end = len(alarmLogs)
	}
	data["logs"] = alarmLogs[start:end]
	response(ctx, 200, data)
}
func (ctrl *Log) GetUserLogs(ctx *gin.Context) {
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
	userLogs, err := ctrl.svc.FindUserLogs(userID, map[string]interface{}{})
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	// 对用户列表进行分页处理
	data := make(map[string]interface{})
	start := (page - 1) * pageSize
	end := start + pageSize
	pages := (len(userLogs) + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}
	data["pages"] = pages
	data["total"] = len(userLogs)
	if start >= len(userLogs) {
		// 响应最后一页
		start = (pages - 1) * pageSize
		end = len(userLogs)
	}
	if end > len(userLogs) {
		end = len(userLogs)
	}
	data["logs"] = userLogs[start:end]
	response(ctx, 200, data)
}

func (ctrl *Log) SelectUserLogs(ctx *gin.Context) {
	form, err := forms.NewUserLogSelect(ctx)
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
	userLogs, err := ctrl.svc.FindUserLogs(userID, form.Conditions)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	// 对用户列表进行分页处理
	data := make(map[string]interface{})
	start := (page - 1) * pageSize
	end := start + pageSize
	pages := (len(userLogs) + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}
	data["pages"] = pages
	data["total"] = len(userLogs)
	if start >= len(userLogs) {
		// 响应最后一页
		start = (pages - 1) * pageSize
		end = len(userLogs)
	}
	if end > len(userLogs) {
		end = len(userLogs)
	}
	data["logs"] = userLogs[start:end]
	response(ctx, 200, data)
}

func (ctrl *Log) GetAlarmLogByID(ctx *gin.Context) {
	alarmLogID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	if alarmLogID <= 0 {
		response(ctx, 40002, nil)
		return
	}
	alarmLog, err := ctrl.svc.GetAlarmLogByID(alarmLogID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if alarmLog == nil {
		response(ctx, 404, nil)
		return
	}
	response(ctx, 200, alarmLog)
}

func (ctrl *Log) GetAlarmLogInfo(ctx *gin.Context) {
	form, err := forms.NewAlarmLogInfoSelect(ctx)
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
	form.Conditions["state"] = 1
	declineCount, err := ctrl.svc.CountAlarmLog(form.Conditions)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	form.Conditions["state"] = 3
	recoverCount, err := ctrl.svc.CountAlarmLog(form.Conditions)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, map[string]interface{}{
		"declineCount": declineCount,
		"recoverCount": recoverCount,
	})
}
