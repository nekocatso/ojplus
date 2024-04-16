package forms

import (
	"github.com/gin-gonic/gin"
)

type AlarmLogSelect struct {
	Page       int `validate:"required,gt=0"`
	PageSize   int `validate:"required,gt=0,lte=100"`
	Query      *AlarmLogConditions
	Conditions map[string]interface{} `validate:"-"`
}

type AlarmLogConditions struct {
	AssetID         int    `validate:"omitempty"`
	RuleID          int    `validate:"omitempty"`
	RuleType        string `validate:"omitempty"`
	State           int    `validate:"omitempty,oneof=1 2 3"`
	Admin           int    `validate:"omitempty"`
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
	if form.Query.Admin != 0 {
		form.Conditions["admin"] = form.Query.Admin
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

type UserLogSelect struct {
	Page       int `validate:"required,gt=0"`
	PageSize   int `validate:"required,gt=0,lte=100"`
	Query      *UserLogConditions
	Conditions map[string]interface{} `validate:"-"`
}

type UserLogConditions struct {
	Username        string `validate:"omitempty"`
	Module          string `validate:"omitempty"`
	Type            string `validate:"omitempty"`
	Phone           string `validate:"omitempty"`
	IP              string `validate:"omitempty"`
	CreateTimeBegin int    `validate:"required_with=CreateTimeEnd,gte=0"`
	CreateTimeEnd   int    `validate:"required_with=CreateTimeBegin,gtefield=CreateTimeBegin"`
}

func NewUserLogSelect(ctx *gin.Context) (*UserLogSelect, error) {
	var form *UserLogSelect
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	if form.Query == nil {
		form.Query = &UserLogConditions{}
	}
	form.Conditions = make(map[string]interface{})
	if form.Query.Username != "" {
		form.Conditions["username"] = form.Query.Username
	}
	if form.Query.Module != "" {
		form.Conditions["module"] = form.Query.Module
	}
	if form.Query.Type != "" {
		form.Conditions["type"] = form.Query.Type
	}
	if form.Query.Phone != "" {
		form.Conditions["phone"] = form.Query.Phone
	}
	if form.Query.IP != "" {
		form.Conditions["ip"] = form.Query.IP
	}
	if form.Query.CreateTimeBegin != 0 {
		form.Conditions["createTimeBegin"] = form.Query.CreateTimeBegin
	}
	if form.Query.CreateTimeEnd != 0 {
		form.Conditions["createTimeEnd"] = form.Query.CreateTimeEnd
	}
	return form, nil
}

func (form *UserLogSelect) check() map[string]string {
	result := make(map[string]string)
	return result
}
