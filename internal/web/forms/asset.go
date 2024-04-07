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
	return result
}

func checkAddress(result map[string]string, typeStr, address string) {
	if strings.EqualFold(typeStr, "ip") {
		reg, _ := regexp2.Compile(`^((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})(\\.((2(5[0-5]|[0-4]\\d))|[0-1]?\\d{1,2})){3}$`, 0)
		matched, _ := reg.FindStringMatch(address)
		if matched != nil {
			result["address"] = "IP地址格式不正确"
		}
	}
	if strings.EqualFold(typeStr, "dns") {
		reg, _ := regexp2.Compile(`^(?=^.{3,255}$)[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+$`, 0)
		matched, _ := reg.FindStringMatch(address)
		if matched != nil {
			result["address"] = "DNS地址格式不正确"
		}
	}
}
