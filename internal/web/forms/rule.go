package forms

import (
	"Alarm/internal/web/models"

	"github.com/gin-gonic/gin"
)

type typeInfo struct {
	Mode         int    `validate:"required_with=LatencyLimit LostLimit"`
	LatencyLimit int    `validate:"omitempty,gt=0"`
	LostLimit    int    `validate:"omitempty,gt=0"`
	EnablePorts  string `validate:"omitempty,max=128"`
	DisablePorts string `validate:"omitempty,max=128"`
}

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
		AlarmID:      form.AlarmID,
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

type RuleUpdate struct {
	Name         *string `validate:"omitempty,max=24"`
	Note         *string `validate:"omitempty,max=256"`
	Assets       []int   `validate:"omitempty"`
	AlarmID      int     `validate:"omitempty"`
	Overtime     int     `validate:"omitempty,gt=100"`
	Interval     int     `validate:"omitempty,gte=5"`
	DeclineLimit int     `validate:"omitempty"`
	RecoverLimit int     `validate:"omitempty"`
	Info         *typeInfo

	UpdateMap     map[string]interface{} `validate:"-"`
	PingUpdateMap map[string]interface{} `validate:"-"`
	TCPUpdateMap  map[string]interface{} `validate:"-"`
}

func NewRuleUpdate(ctx *gin.Context) (*RuleUpdate, error) {
	var form *RuleUpdate
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}

	form.UpdateMap = make(map[string]interface{})
	form.PingUpdateMap = make(map[string]interface{})
	form.TCPUpdateMap = make(map[string]interface{})

	if form.Info != nil {
		if form.Info.Mode != 0 {
			form.PingUpdateMap["mode"] = form.Info.Mode
		}
		if form.Info.LatencyLimit != 0 {
			form.PingUpdateMap["latency_limit"] = form.Info.LatencyLimit
		}
		if form.Info.LostLimit != 0 {
			form.PingUpdateMap["lost_limit"] = form.Info.LostLimit
		}
	}
	if form.Info != nil {
		if len(form.Info.EnablePorts) > 0 || len(form.Info.DisablePorts) > 0 {
			form.TCPUpdateMap["enable_ports"] = form.Info.EnablePorts
			form.TCPUpdateMap["disable_ports"] = form.Info.DisablePorts
		}
	}
	if form.Name != nil {
		form.UpdateMap["name"] = *form.Name
	}
	if form.Note != nil {
		form.UpdateMap["note"] = *form.Note
	}
	if form.AlarmID != 0 {
		form.UpdateMap["alarm_id"] = form.AlarmID
	}
	if form.Overtime != 0 {
		form.UpdateMap["overtime"] = form.Overtime
	}
	if form.Interval != 0 {
		form.UpdateMap["interval"] = form.Interval
	}
	if form.DeclineLimit != 0 {
		form.UpdateMap["decline_limit"] = form.DeclineLimit
	}
	if form.RecoverLimit != 0 {
		form.UpdateMap["recover_limit"] = form.RecoverLimit
	}
	return form, nil
}

func (form *RuleUpdate) check() map[string]string {
	result := make(map[string]string)
	if form.Info != nil && form.Info.LatencyLimit >= form.Overtime {
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
