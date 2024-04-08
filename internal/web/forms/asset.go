package forms

import (
	"Alarm/internal/web/models"
	"strings"

	"github.com/dlclark/regexp2"

	"github.com/gin-gonic/gin"
)

type AssetCreate struct {
	Name    string `validate:"required,max=24"`
	Type    string `validate:"required,max=12"`
	Address string `validate:"required,max=128"`
	Note    string `validate:"omitempty,max=128"`
	Enable  bool   `validate:"omitempty"`
	Users   []int  `validate:"omitempty"`
	Rules   []int  `validate:"omitempty"`
	Model   *models.Asset
}

func NewAssetCreate(ctx *gin.Context) (*AssetCreate, error) {
	var form *AssetCreate
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	var state int
	if form.Enable {
		state = 3
	} else {
		state = -1
	}
	form.Model = &models.Asset{
		Name:    form.Name,
		Address: form.Address,
		State:   state,
		Type:    form.Type,
		Note:    form.Note,
	}
	return form, nil
}

func (form *AssetCreate) check() map[string]string {
	result := make(map[string]string)
	if form.Enable && len(form.Rules) == 0 {
		result["state"] = "未绑定规则时无法启用监测"
	}
	checkAddress(result, form.Type, form.Address)
	checkType(result, form.Type)
	return result
}

func checkType(result map[string]string, t string) {
	if t != "ip" && t != "domain" {
		result["type"] = "非可选的类型(ip, domain)"
	}
}
func checkAddress(result map[string]string, typeStr, address string) {
	if strings.EqualFold(typeStr, "ip") {
		reg, _ := regexp2.Compile(`^((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})(\\.((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})){3}$`, 0)
		matched, _ := reg.FindStringMatch(address)
		if matched != nil {
			result["address"] = "IP地址格式不正确"
		}
	}
	if strings.EqualFold(typeStr, "domain") {
		reg, _ := regexp2.Compile(`^\.([a-zA-Z]+)$`, 0)
		matched, _ := reg.FindStringMatch(address)
		if matched != nil {
			result["address"] = "domain地址格式不正确"
		}
	}
}

type AssetSelect struct {
	Page       int `validate:"required,gt=0"`
	PageSize   int `validate:"required,gt=0,lte=100"`
	Query      *AssetConditions
	Model      *models.Asset
	Conditions map[string]interface{}
}

type AssetConditions struct {
	Name            string `validate:"omitempty"`
	Type            string `validate:"omitempty"`
	CreatorID       int    `validate:"omitempty,gt=0"`
	State           int    `validate:"omitempty,gte=-1,lte=3"`
	Enable          int    `validate:"omitempty"`
	CreateTimeBegin int    `validate:"required_with=CreateTimeEnd"`
	CreateTimeEnd   int    `validate:"required_with=CreateTimeBegin,gtefield=CreateTimeBegin"`
}

func NewAssetSelect(ctx *gin.Context) (*AssetSelect, error) {
	var form *AssetSelect
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	var state int
	if form.Query == nil {
		form.Query = &AssetConditions{}
	}
	if form.Query.State != 0 {
		if form.Query.Enable > 0 {
			state = 3
		} else {
			state = -1
		}
	} else {
		state = form.Query.State
	}
	form.Model = &models.Asset{
		Name:      form.Query.Name,
		State:     state,
		Type:      form.Query.Type,
		CreatorID: form.Query.CreatorID,
	}
	form.Conditions = make(map[string]interface{})
	if form.Query.Name != "" {
		form.Conditions["name"] = form.Query.Name
	}
	if form.Query.Type != "" {
		form.Conditions["type"] = form.Query.Type
	}
	if form.Query.CreatorID != 0 {
		form.Conditions["creatorID"] = form.Query.CreatorID
	}
	if form.Query.Enable != 0 {
		form.Conditions["enable"] = form.Query.Enable
	}
	if form.Query.State != 0 {
		form.Conditions["state"] = form.Query.State
	}
	if form.Query.CreateTimeBegin != 0 {
		form.Conditions["createTimeBegin"] = form.Query.CreateTimeBegin
	}
	if form.Query.CreateTimeEnd != 0 {
		form.Conditions["createTimeEnd"] = form.Query.CreateTimeEnd
	}
	return form, nil
}

func (form *AssetSelect) check() map[string]string {
	result := make(map[string]string)
	if form.Query.Enable*form.Query.State < 0 {
		result["state"] = "state与enable相冲突"
	}
	return result
}
