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

type Alarm struct {
	svc    *services.Alarm
	cfg    map[string]interface{}
	logger *logs.Logger
}

func NewAlarm(cfg map[string]interface{}) *Alarm {
	svc := services.NewAlarm(cfg)
	logger := logs.NewLogger(cfg["db"].(*models.Database))
	return &Alarm{svc: svc, cfg: cfg, logger: logger}
}

func (ctrl *Alarm) CreateAlarm(ctx *gin.Context) {
	form, err := forms.NewAlarmCreate(ctx)
	if err != nil {
		log.Println(err)
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
	alarm := form.Model
	userID := GetUserIDByContext(ctx)
	user, err := ctrl.svc.GetUserByID(userID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if user == nil {
		response(ctx, 404, nil)
		return
	}
	alarm.CreatorID = userID
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
	err = ctrl.logger.SaveUserLog(ctx, user, &logs.UserLog{
		Module:  "通知策略",
		Type:    "新增",
		Content: alarm.Name,
	})
	if err != nil {
		log.Println(err)
	}
}

func (ctrl *Alarm) UpdateAlarmByID(ctx *gin.Context) {
	// 数据校验
	assetID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	form, err := forms.NewAlarmUpdate(ctx)
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
	// 权限校验
	userID := GetUserIDByContext(ctx)
	user, err := ctrl.svc.GetUserByID(userID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if user == nil {
		response(ctx, 404, nil)
		return
	}
	alarm, err := ctrl.svc.GetAlarmByID(assetID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if alarm == nil {
		response(ctx, 404, nil)
		return
	}
	if alarm.CreatorID != userID {
		response(ctx, 404, nil)
		return
	}
	// 更新数据
	err = ctrl.svc.UpdateAlarm(assetID, form.UpdateMap)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, nil)
	err = ctrl.logger.SaveUserLog(ctx, user, &logs.UserLog{
		Module:  "通知策略",
		Type:    "编辑",
		Content: alarm.Name,
	})
	if err != nil {
		log.Println(err)
	}
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
	alarms, err := ctrl.svc.FindAlarms(userID, form.Conditions)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	// 对用户列表进行分页处理
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

func (ctrl *Alarm) GetAlarmByID(ctx *gin.Context) {
	alarmIDStr := ctx.Param("id")
	alarmID, err := strconv.Atoi(alarmIDStr)
	if err != nil {
		response(ctx, 40002, nil)
		return
	}
	userID := GetUserIDByContext(ctx)
	access, err := ctrl.svc.IsAccessAlarm(alarmID, userID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if !access {
		response(ctx, 404, nil)
		return
	}
	alarm, err := ctrl.svc.GetAlarmByID(alarmID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, alarm)
}

func (ctrl *Alarm) DeleteAlarmByID(ctx *gin.Context) {
	// 数据校验
	alarmID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	if alarmID <= 0 {
		response(ctx, 40002, nil)
		return
	}
	// 权限校验
	userID := GetUserIDByContext(ctx)
	user, err := ctrl.svc.GetUserByID(userID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if user == nil {
		response(ctx, 404, nil)
		return
	}
	alarm, err := ctrl.svc.GetAlarmByID(alarmID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if alarm == nil {
		response(ctx, 404, nil)
		return
	}
	if alarm.CreatorID != userID {
		response(ctx, 404, nil)
		return
	}
	// 数据处理
	err = ctrl.svc.DeleteAlarmByID(alarmID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, nil)
	err = ctrl.logger.SaveUserLog(ctx, user, &logs.UserLog{
		Module:  "通知策略",
		Type:    "删除",
		Content: alarm.Name,
	})
	if err != nil {
		log.Println(err)
	}
}
