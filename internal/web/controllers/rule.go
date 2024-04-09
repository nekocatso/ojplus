package controllers

import (
	"Alarm/internal/web/forms"
	"Alarm/internal/web/services"
	"log"

	"github.com/gin-gonic/gin"
)

type Rule struct {
	svc *services.Rule
	cfg map[string]interface{}
}

func NewRule(cfg map[string]interface{}) *Rule {
	svc := services.NewRule(cfg)
	return &Rule{svc: svc, cfg: cfg}
}

func (ctrl *Rule) CreateRule(ctx *gin.Context) {
	ruleForm, err := forms.NewRuleCreate(ctx)
	if err != nil {
		log.Println(err)
		response(ctx, 40001, nil)
		return
	}
	isValid, errorsMap, err := forms.Verify(ruleForm)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if !isValid {
		response(ctx, 40002, errorsMap)
		return
	}
	rule := ruleForm.Model
	if rule.Type == "ping" {
		err = ctrl.svc.CreatePingRule(rule, ruleForm.PingInfo)
		if err != nil {
			log.Println(err)
			response(ctx, 500, nil)
			return
		}
	} else if rule.Type == "tcp" {
		err = ctrl.svc.CreateTCPRule(rule, ruleForm.TCPInfo)
		if err != nil {
			log.Println(err)
			response(ctx, 500, nil)
			return
		}
	} else {
		response(ctx, 40002, nil)
		return
	}
	if ruleForm.Assets != nil {
		err := ctrl.svc.BindAssets(rule.ID, ruleForm.Assets)
		if err != nil {
			response(ctx, 400, nil)
			return
		}
	}
	response(ctx, 201, map[string]interface{}{"ruleID": rule.ID})
}
