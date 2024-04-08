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

func (ctrl *Asset) FindAsset(ctx *gin.Context) {
	// pageStr := ctx.Query("page")
	// pageSizeStr := ctx.Query("pageSize")
	// var page, pageSize int
	// if pageStr != "" && pageSizeStr != "" {
	// 	page, err := strconv.Atoi(pageStr)
	// 	if err != nil || page <= 0 {
	// 		response(ctx, 40002, nil)
	// 		return
	// 	}
	// 	if pageSize > 100 {
	// 		response(ctx, 40003, nil)
	// 		return
	// 	}
	// 	pageSize, err := strconv.Atoi(pageSizeStr)
	// 	if err != nil || pageSize <= 0 {
	// 		response(ctx, 40002, nil)
	// 		return
	// 	}
	// } else {
	// 	page = 1
	// 	pageSize = 10
	// }
}
