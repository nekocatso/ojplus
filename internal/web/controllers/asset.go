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
	has, hasMessage, err := ctrl.svc.GetAssetExistInfo(asset)
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
		err := ctrl.svc.BindUsers(asset.ID, userIDs)
		if err != nil {
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
	// 数据校验
	assetID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	has, err := ctrl.svc.IsAssetExistByID(assetID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if !has {
		response(ctx, 404, nil)
		return
	}
	form, err := forms.NewAssetUpdate(ctx)
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
	// 权限校验
	claims := ctx.Value("claims").(jwt.MapClaims)
	userID := claims["userID"].(int)
	access, err := ctrl.svc.IsAccessAsset(asset.ID, userID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if !access {
		response(ctx, 404, nil)
		return
	}
	// 更新数据
	err = ctrl.svc.UpdateAsset(asset)
	if err != nil {
		response(ctx, 500, nil)
		return
	}
}

func (ctrl *Asset) GetAssets(ctx *gin.Context) {
	// 数据校验
	pageStr := ctx.Query("page")
	pageSizeStr := ctx.Query("pageSize")
	var page, pageSize int
	if pageStr != "" && pageSizeStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page <= 0 {
			response(ctx, 40002, nil)
			return
		}
		if pageSize > 100 {
			response(ctx, 40003, nil)
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
	claims := ctx.Value("claims").(jwt.MapClaims)
	userID := claims["userID"].(int)
	assets, err := ctrl.svc.FindAssets(userID, map[string]interface{}{})
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
		start = (pages - 1) * pageSize
		end = len(assets)
	}
	if end > len(assets) {
		end = len(assets)
	}
	data["assets"] = assets[start:end]
	response(ctx, 200, data)
}

func (ctrl *Asset) GetAssetByID(ctx *gin.Context) {
	// 数据校验
	assetID, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	has, err := ctrl.svc.IsAssetExistByID(assetID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if !has {
		response(ctx, 404, nil)
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
	// 获取资产信息
	asset, err := ctrl.svc.GetAssetByID(assetID)
	if err != nil {
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, asset)
}

func (ctrl *Asset) SelectAssets(ctx *gin.Context) {
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
	assets, err := ctrl.svc.FindAssets(userID, form.Conditions)
	if err != nil {
		log.Println(err)
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

func (ctrl *Asset) GetAssetIDs(ctx *gin.Context) {
	claims := ctx.Value("claims").(jwt.MapClaims)
	userID := claims["userID"].(int)
	assets, err := ctrl.svc.FindAssets(userID, map[string]interface{}{})
	if err != nil {
		response(ctx, 500, nil)
		return
	}
	assetIDs := []int{}
	for _, asset := range assets {
		assetIDs = append(assetIDs, asset.ID)
	}
	response(ctx, 200, assetIDs)
}
