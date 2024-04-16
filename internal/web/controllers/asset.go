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

type Asset struct {
	svc    *services.Asset
	cfg    map[string]interface{}
	logger *logs.Logger
}

func NewAsset(cfg map[string]interface{}) *Asset {
	svc := services.NewAsset(cfg)
	logger := logs.NewLogger(cfg["db"].(*models.Database))
	return &Asset{svc: svc, cfg: cfg, logger: logger}
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
	userID := GetUserIDByContext(ctx)
	asset.CreatorID = userID
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
	err = ctrl.svc.CreateAsset(asset, form.Users, form.Rules)
	if err != nil {
		log.Println(err)
		response(ctx, 404, nil)
		return
	}
	if merr, ok := err.(*mysql.MySQLError); ok {
		if merr.Number == 1062 {
			response(ctx, 40901, nil)
			return
		}
	}
	response(ctx, 201, map[string]int{"assetID": asset.ID})
	err = ctrl.logger.SaveUserLog(ctx, userID, &logs.UserLog{
		Module:  "资产管理",
		Type:    "新增",
		Content: asset.Name,
	})
	if err != nil {
		log.Println(err)
	}
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
	// 更新数据
	err = ctrl.svc.UpdateAsset(assetID, form.UpdateMap, form.Users, form.Rules)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, nil)
	asset, err := ctrl.svc.GetAssetByID(assetID)
	if err != nil {
		log.Println(err)
		return
	}
	err = ctrl.logger.SaveUserLog(ctx, userID, &logs.UserLog{
		Module:  "资产管理",
		Type:    "编辑",
		Content: asset.Name,
	})
	if err != nil {
		log.Println(err)
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
	userID := GetUserIDByContext(ctx)
	access, err := ctrl.svc.IsAccessAsset(assetID, userID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	if !access {
		log.Println("")
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
	userID := GetUserIDByContext(ctx)
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
	userID := GetUserIDByContext(ctx)
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

func (ctrl *Asset) GetAssetsByRuleID(ctx *gin.Context) {
	ruleID, err := strconv.Atoi(ctx.Param("ruleID"))
	if err != nil {
		response(ctx, 40001, nil)
		return
	}
	userID := GetUserIDByContext(ctx)
	assets, err := ctrl.svc.FindAssets(userID, map[string]interface{}{
		"ruleID": ruleID,
	})
	if err != nil {
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, assets)
}

func (ctrl *Asset) DeleteAsset(ctx *gin.Context) {
	// 数据校验
	assetID, err := strconv.Atoi(ctx.Param("id"))
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
	// 数据处理
	err = ctrl.svc.DeleteAsset(assetID)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, nil)
}

func (ctrl *Asset) GetAssetsInfo(ctx *gin.Context) {
	userID := GetUserIDByContext(ctx)
	conditions := make(map[string]interface{})
	assetCount, err := ctrl.svc.CountAsset(userID, conditions)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	conditions["enable"] = true
	assetEnableCount, err := ctrl.svc.CountAsset(userID, conditions)
	if err != nil {
		log.Println(err)
		response(ctx, 500, nil)
		return
	}
	response(ctx, 200, map[string]interface{}{
		"assetCount":  assetCount,
		"enableCount": assetEnableCount,
	})
}
