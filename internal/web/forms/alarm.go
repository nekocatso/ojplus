package forms

import (
	"Alarm/internal/web/models"

	"github.com/gin-gonic/gin"
)

type AlarmCreate struct {
	Name     string                `validate:"required,max=24"`
	Note     *string               `validate:"omitempty,max=256"`
	Interval int                   `validate:"omitempty"`
	Mails    []string              `validate:"required"`
	Model    *models.AlarmTemplate `validate:"-"`
}

func NewAlarmCreate(ctx *gin.Context) (*AlarmCreate, error) {
	var form *AlarmCreate
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	form.Model = &models.AlarmTemplate{
		Name:  form.Name,
		Note:  form.Note,
		Mails: form.Mails,
	}
	return form, nil
}

func (form *AlarmCreate) check() map[string]string {
	result := make(map[string]string)
	if len(form.Mails) == 0 {
		result["mails"] = "邮件列表不能为空"
	}
	return result
}

type AlarmUpdate struct {
	Name      *string                `validate:"omitempty,max=24"`
	Note      *string                `validate:"omitempty,max=256"`
	Interval  int                    `validate:"omitempty"`
	Mails     []string               `validate:"omitempty"`
	UpdateMap map[string]interface{} `validate:"-"`
}

func NewAlarmUpdate(ctx *gin.Context) (*AlarmUpdate, error) {
	var form *AlarmUpdate
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}

	form.UpdateMap = make(map[string]interface{})
	if form.Name != nil {
		form.UpdateMap["name"] = *form.Name
	}
	if form.Note != nil {
		form.UpdateMap["note"] = *form.Note
	}
	if form.Interval != 0 {
		form.UpdateMap["interval"] = form.Interval
	}
	if form.Mails != nil {
		form.UpdateMap["mails"] = form.Mails
	}

	return form, nil
}

func (form *AlarmUpdate) check() map[string]string {
	result := make(map[string]string)
	if form.Mails != nil && len(form.Mails) == 0 {
		result["mails"] = "邮件列表不能为空"
	}
	return result
}

type AlarmTemplateSelect struct {
	Page       int `validate:"required,gt=0"`
	PageSize   int `validate:"required,gt=0,lte=100"`
	Query      *AlarmTemplateConditions
	Model      *models.AlarmTemplate  `validate:"-"`
	Conditions map[string]interface{} `validate:"-"`
}

type AlarmTemplateConditions struct {
	Name            string `validate:"omitempty"`
	RuleID          int    `validate:"omitempty"`
	CreateTimeBegin int    `validate:"required_with=CreateTimeEnd,gte=0"`
	CreateTimeEnd   int    `validate:"required_with=CreateTimeBegin,gtefield=CreateTimeBegin"`
}

func NewAlarmSelect(ctx *gin.Context) (*AlarmTemplateSelect, error) {
	var form *AlarmTemplateSelect
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	if form.Query == nil {
		form.Query = &AlarmTemplateConditions{}
	}
	form.Model = &models.AlarmTemplate{
		Name: form.Query.Name,
	}
	form.Conditions = make(map[string]interface{})
	if form.Query.Name != "" {
		form.Conditions["name"] = form.Query.Name
	}
	if form.Query.RuleID != 0 {
		form.Conditions["ruleID"] = form.Query.RuleID
	}
	if form.Query.CreateTimeBegin != 0 {
		form.Conditions["createTimeBegin"] = form.Query.CreateTimeBegin
	}
	if form.Query.CreateTimeEnd != 0 {
		form.Conditions["createTimeEnd"] = form.Query.CreateTimeEnd
	}
	return form, nil
}

func (form *AlarmTemplateSelect) check() map[string]string {
	result := make(map[string]string)
	return result
}
