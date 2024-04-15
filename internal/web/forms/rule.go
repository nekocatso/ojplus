package forms

import (
	"Alarm/internal/web/models"

	"github.com/gin-gonic/gin"
)

type RuleCreate struct {
	Name         string  `validate:"required,max=24"`
	Type         string  `validate:"required,max=12"`
	Note         *string `validate:"omitempty,max=256"`
	Assets       []int   `validate:"omitempty"`
	AlarmID      int     `validate:"omitempty"`
	Overtime     int     `validate:"required,gt=100"`
	Interval     int     `validate:"required,gte=5"`
	DeclineLimit int     `validate:"required"`
	RecoverLimit int     `validate:"required"`
	Info         *typeInfo

	Model    *models.Rule     `validate:"-"`
	PingInfo *models.PingInfo `validate:"-"`
	TCPInfo  *models.TCPInfo  `validate:"-"`
}

type typeInfo struct {
	Mode         int    `validate:"required_with=LatencyLimit LostLimit"`
	LatencyLimit int    `validate:"omitempty,gt=0"`
	LostLimit    int    `validate:"omitempty,gt=0"`
	EnablePorts  string `validate:"omitempty,max=128"`
	DisablePorts string `validate:"omitempty,max=128"`
}

func NewRuleCreate(ctx *gin.Context) (*RuleCreate, error) {
	var form *RuleCreate
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	form.Model = &models.Rule{
		Name:         form.Name,
		Type:         form.Type,
		Overtime:     form.Overtime,
		Interval:     form.Interval,
		DeclineLimit: form.DeclineLimit,
		RecoverLimit: form.RecoverLimit,
		Note:         form.Note,
	}
	form.PingInfo = &models.PingInfo{
		Mode:         form.Info.Mode,
		LatencyLimit: form.Info.LatencyLimit,
		LostLimit:    form.Info.LostLimit,
	}
	form.TCPInfo = &models.TCPInfo{
		EnablePorts:  form.Info.EnablePorts,
		DisablePorts: form.Info.DisablePorts,
	}
	return form, nil

}

func (form *RuleCreate) check() map[string]string {
	result := make(map[string]string)
	if form.Info.LatencyLimit >= form.Overtime {
		result["latencyLimit"] = "latencyLimit必须小于overtime"
	}
	return result
}

type RuleSelect struct {
	Page       int `validate:"required,gt=0"`
	PageSize   int `validate:"required,gt=0,lte=100"`
	Query      *RuleConditions
	Model      *models.Rule           `validate:"-"`
	Conditions map[string]interface{} `validate:"-"`
}

type RuleConditions struct {
	Name            string `validate:"omitempty"`
	Type            string `validate:"omitempty"`
	CreatorID       int    `validate:"omitempty,gt=0"`
	AssetID         int    `validate:"omitempty,gt=0"`
	CreateTimeBegin int    `validate:"required_with=CreateTimeEnd,gte=0"`
	CreateTimeEnd   int    `validate:"required_with=CreateTimeBegin,gtefield=CreateTimeBegin"`
}

func NewRuleSelect(ctx *gin.Context) (*RuleSelect, error) {
	var form *RuleSelect
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	if form.Query == nil {
		form.Query = &RuleConditions{}
	}
	form.Model = &models.Rule{
		Name:      form.Query.Name,
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
	if form.Query.AssetID != 0 {
		form.Conditions["assetID"] = form.Query.AssetID
	}
	if form.Query.CreateTimeBegin != 0 {
		form.Conditions["createTimeBegin"] = form.Query.CreateTimeBegin
	}
	if form.Query.CreateTimeEnd != 0 {
		form.Conditions["createTimeEnd"] = form.Query.CreateTimeEnd
	}
	return form, nil
}

func (form *RuleSelect) check() map[string]string {
	result := make(map[string]string)
	return result
}
