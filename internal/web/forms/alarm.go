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
