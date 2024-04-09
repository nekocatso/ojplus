package forms

import (
	"Alarm/internal/web/models"

	"github.com/gin-gonic/gin"
)

type RuleCreate struct {
	Name         string       `validate:"required,max=24"`
	Type         string       `validate:"required,max=12"`
	AlarmID      int          `validate:"required"`
	Overtime     int          `validate:"required"`
	DeclineLimit int          `validate:"required"`
	RecoverLimit int          `validate:"required"`
	Model        *models.Rule `validate:"-"`
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
		AlarmID:      form.AlarmID,
		Overtime:     form.Overtime,
		DeclineLimit: form.DeclineLimit,
		RecoverLimit: form.RecoverLimit,
	}
	return form, nil
}

func (form *RuleCreate) check() map[string]string {
	result := make(map[string]string)
	return result
}

type PingCreate struct {
	Mode         int              `validate:"required"`
	LatencyLimit int              `validate:"required"`
	LostLimit    int              `validate:"required"`
	Model        *models.PingInfo `validate:"-"`
}

func NewPingCreate(ctx *gin.Context) (*PingCreate, error) {
	var form *PingCreate
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	form.Model = &models.PingInfo{
		Mode:         form.Mode,
		LatencyLimit: form.LatencyLimit,
		LostLimit:    form.LostLimit,
	}
	return form, nil
}
