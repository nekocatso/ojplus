package controllers

import (
	"Alarm/internal/web/forms"
	"Alarm/internal/web/services"
	"log"

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
