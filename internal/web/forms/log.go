package forms

import (
	"Alarm/internal/web/models"

	"github.com/gin-gonic/gin"
)

type AlarmLogSelect struct {
	Page       int `validate:"required,gt=0"`
	PageSize   int `validate:"required,gt=0,lte=100"`
	Query      *AlarmLogConditions
	Model      *models.AlarmLog       `validate:"-"`
	Conditions map[string]interface{} `validate:"-"`
}

type AlarmLogConditions struct {
	AssetID         int    `validate:"omitempty"`
	RuleID          int    `validate:"omitempty"`
	RuleType        string `validate:"omitempty"`
	State           int    `validate:"omitempty,oneof=1 2 3"`
	AssetCreator    int    `validate:"omitempty"`
	CreateTimeBegin int    `validate:"required_with=CreateTimeEnd,gte=0"`
	CreateTimeEnd   int    `validate:"required_with=CreateTimeBegin,gtefield=CreateTimeBegin"`
}

func NewAlarmLogSelect(ctx *gin.Context) (*AlarmLogSelect, error) {
	var form *AlarmLogSelect
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	if form.Query == nil {
		form.Query = &AlarmLogConditions{}
	}
	form.Model = &models.AlarmLog{
		AssetID: form.Query.AssetID,
		RuleID:  form.Query.RuleID,
		State:   form.Query.State,
	}
	form.Conditions = make(map[string]interface{})
	if form.Query.AssetID != 0 {
		form.Conditions["assetID"] = form.Query.AssetID
	}
	if form.Query.RuleID != 0 {
		form.Conditions["ruleID"] = form.Query.RuleID
	}
	if form.Query.RuleType != "" {
		form.Conditions["ruleType"] = form.Query.RuleType
	}
	if form.Query.State != 0 {
		form.Conditions["state"] = form.Query.State
	}
	if form.Query.AssetCreator != 0 {
		form.Conditions["assetCreator"] = form.Query.AssetCreator
	}
	if form.Query.CreateTimeBegin != 0 {
		form.Conditions["createTimeBegin"] = form.Query.CreateTimeBegin
	}
	if form.Query.CreateTimeEnd != 0 {
		form.Conditions["createTimeEnd"] = form.Query.CreateTimeEnd
	}
	return form, nil
}

func (form *AlarmLogSelect) check() map[string]string {
	result := make(map[string]string)
	return result
}
