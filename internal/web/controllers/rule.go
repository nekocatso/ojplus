package controllers

import (
	"Alarm/internal/web/forms"
	"Alarm/internal/web/services"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
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
	claims := ctx.Value("claims").(jwt.MapClaims)
	userID := claims["userID"].(int)
	rule.CreatorID = userID
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
		err := ctrl.svc.BindAssets(rule.ID, ruleForm.Assets)
		if err != nil {
			response(ctx, 400, nil)
			return
		}
	}
	response(ctx, 201, map[string]interface{}{"ruleID": rule.ID})
}

func (ctrl *Rule) GetRules(ctx *gin.Context) {
	claims := ctx.Value("claims").(jwt.MapClaims)
	userID := claims["userID"].(int)
	rules, err := ctrl.svc.FindRules(userID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, rules)
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
	claims := ctx.Value("claims").(jwt.MapClaims)
	userID := claims["userID"].(int)
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
