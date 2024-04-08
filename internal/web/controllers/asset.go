package controllers

import (
	"Alarm/internal/web/forms"
	"Alarm/internal/web/services"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v5"
)

type Asset struct {
	svc *services.Asset
	cfg map[string]interface{}
}

func NewAsset(cfg map[string]interface{}) *Asset {
	svc := services.NewAsset(cfg)
	return &Asset{svc: svc, cfg: cfg}
}

func (ctrl *Asset) CreateAsset(ctx *gin.Context) {
	// 数据校验
	form, err := forms.NewAssetCreate(ctx)
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
	asset := form.Model
	claims := ctx.Value("claims").(jwt.MapClaims)
	asset.CreatorID = claims["userID"].(int)
	has, hasMessage, err := ctrl.svc.IsAssetExist(asset)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if has {
		responseWithMessage(ctx, hasMessage, 40901, nil)
		return
	}
	// 创建资产
	err = ctrl.svc.CreateAsset(asset)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if merr, ok := err.(*mysql.MySQLError); ok {
		if merr.Number == 1062 {
			response(ctx, 40901, nil)
			return
		}
	}
	// 绑定用户或规则
	err = ctrl.svc.GetAssetInfo(asset)
	if err != nil || asset.ID == 0 {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if form.Users != nil {
		userIDs := append(form.Users, asset.CreatorID)
		err = ctrl.svc.BindUsers(asset.ID, userIDs)
		if merr, ok := err.(*mysql.MySQLError); ok {
			if merr.Number == 1062 {
				response(ctx, 40901, nil)
				return
			}
		}
		if err != nil || asset.ID == 0 {
			log.Println(err)
			response(ctx, 400, nil)
			return
		}
	}
	if form.Rules != nil {
		err = ctrl.svc.BindRules(asset.ID, form.Rules)
		if merr, ok := err.(*mysql.MySQLError); ok {
			if merr.Number == 1062 {
				response(ctx, 40901, nil)
				return
			}
		}
		if err != nil || asset.ID == 0 {
			log.Println(err)
			response(ctx, 400, nil)
			return
		}
	}
	response(ctx, 201, map[string]int{"assetID": asset.ID})
}

func (ctrl *Asset) UpdateAsset(ctx *gin.Context) {

}

func (ctrl *Asset) SelectAsset(ctx *gin.Context) {
	// 数据校验
	form, err := forms.NewAssetSelect(ctx)
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
	claims := ctx.Value("claims").(jwt.MapClaims)
	userID := claims["userID"].(int)
	assets, err := ctrl.svc.QueryAssetsWithConditions(userID, form.Conditions)
	if err != nil {
		response(ctx, 500, nil)
		return
	}
	// 分页处理
	data := make(map[string]interface{})
	start := (page - 1) * pageSize
	end := start + pageSize
	pages := (len(assets) + pageSize - 1) / pageSize
	if pages == 0 {
		pages = 1
	}
	data["pages"] = pages
	data["total"] = len(assets)
	if start >= len(assets) {
		// 响应最后一页
		start = (pages - 1) * pageSize
		end = len(assets)
	} else if end > len(assets) {
		end = len(assets)
	}
	data["assets"] = assets[start:end]
	response(ctx, 200, data)
}
