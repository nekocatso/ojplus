package controllers

import (
	"Alarm/internal/web/forms"
	"Alarm/internal/web/services"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
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
	// 数据校验
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
	userID := GetUserIDByContext(ctx)
	rule.CreatorID = userID
	// 权限校验
	for _, asset := range ruleForm.Assets {
		access, err := ctrl.svc.IsAccessAsset(asset, userID)
		if err != nil {
			response(ctx, 500, nil)
			return
		}
		if !access {
			response(ctx, 404, nil)
			return
		}
	}
	// 创建规则
	if rule.Type == "ping" {
		err = ctrl.svc.CreatePingRule(rule, ruleForm.PingInfo)
	} else if rule.Type == "tcp" {
		err = ctrl.svc.CreateTCPRule(rule, ruleForm.TCPInfo)
	} else {
		response(ctx, 40002, nil)
		return
	}
	if merr, ok := err.(*mysql.MySQLError); ok {
		if merr.Number == 1062 {
			response(ctx, 40901, nil)
			return
		}
	} else if err != nil {
		response(ctx, 500, nil)
		return
	}
	if ruleForm.Assets != nil {
		err := ctrl.svc.BindAssets(rule.ID, ruleForm.Assets, userID)
		if err != nil {
			log.Println(err)
			response(ctx, 400, nil)
			return
		}
	}
	response(ctx, 201, map[string]interface{}{"ruleID": rule.ID})
}

func (ctrl *Rule) GetRules(ctx *gin.Context) {
	userID := GetUserIDByContext(ctx)
	rules, err := ctrl.svc.FindRules(userID, nil)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, rules)
}

func (ctrl *Rule) SelectRules(ctx *gin.Context) {
	// 数据校验
	form, err := forms.NewRuleSelect(ctx)
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
	rules, err := ctrl.svc.FindRules(userID, form.Conditions)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	// 分页处理
	data := make(map[string]interface{})
	start := (page - 1) * pageSize
	end := start + pageSize
	pages := (len(rules) + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}
	data["pages"] = pages
	data["total"] = len(rules)
	if start >= len(rules) {
		// 响应最后一页
		start = (pages - 1) * pageSize
		end = len(rules)
	} else if end > len(rules) {
		end = len(rules)
	}
	data["rules"] = rules[start:end]
	response(ctx, 200, data)
}

func (ctrl *Rule) GetRuleIDsByAssetID(ctx *gin.Context) {
	// 数据校验
	assetID, err := strconv.Atoi(ctx.Param("assetID"))
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	if assetID <= 0 {
		response(ctx, 40002, nil)
		return
	}
	// 权限校验
	userID := GetUserIDByContext(ctx)
	access, err := ctrl.svc.IsAccessAsset(assetID, userID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if !access {
		response(ctx, 404, nil)
		return
	}
	// 获取数据
	var ruleIDs []int
	ruleIDs, err = ctrl.svc.GetRuleIDsByAssetID(assetID)
	if err != nil {
		response(ctx, 500, nil)
		log.Println(err)
		return
	}
	if ruleIDs == nil {
		response(ctx, 404, nil)
		return
	}
	response(ctx, 200, ruleIDs)
}

func (ctrl *Rule) GetRulesByAssetID(ctx *gin.Context) {
	// 数据校验
	assetID, err := strconv.Atoi(ctx.Param("assetID"))
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	if assetID <= 0 {
		response(ctx, 40002, nil)
		return
	}
	// 权限校验
	userID := GetUserIDByContext(ctx)
	access, err := ctrl.svc.IsAccessAsset(assetID, userID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if !access {
		response(ctx, 404, nil)
		return
	}
	rules, err := ctrl.svc.GetRulesByAssetID(assetID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, rules)
}

func (ctrl *Rule) GetRuleByID(ctx *gin.Context) {
	// 数据校验
	ruleID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	if ruleID <= 0 {
		response(ctx, 40002, nil)
		return
	}
	// 权限校验
	userID := GetUserIDByContext(ctx)
	access, err := ctrl.svc.IsAccessRule(ruleID, userID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if !access {
		response(ctx, 404, nil)
		return
	}
	// 获取数据
	rule, err := ctrl.svc.GetRuleByID(ruleID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if rule.Type == "ping" {
		rule.Info, err = ctrl.svc.GetPingInfo(ruleID)
	} else if rule.Type == "tcp" {
		rule.Info, err = ctrl.svc.GetTCPInfo(ruleID)
	} else {
		log.Println("a rule type not ping or tcp")
		response(ctx, 500, nil)
		return
	}
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if rule == nil {
		response(ctx, 404, nil)
		return
	}
	response(ctx, 200, rule)
}
