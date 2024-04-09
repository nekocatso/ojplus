package forms

import (
	"Alarm/internal/web/models"

	"github.com/gin-gonic/gin"
)

type RuleCreate struct {
	Name         string `validate:"required,max=24"`
	Type         string `validate:"required,max=12"`
	Assets       []int  `validate:"omitempty"`
	Overtime     int    `validate:"required"`
	Interval     int    `validate:"required"`
	DeclineLimit int    `validate:"required"`
	RecoverLimit int    `validate:"required"`

	Mode         int `validate:"required_with=LatencyLimit LostLimit"`
	LatencyLimit int `validate:"required_with=Mode LostLimit"`
	LostLimit    int `validate:"required_with=Mode LatencyLimit"`

	EnablePorts  string `validate:"required_with=DisablePorts,max=128"`
	DisablePorts string `validate:"required_with=EnablePorts,max=128"`

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
		DeclineLimit: form.DeclineLimit,
		RecoverLimit: form.RecoverLimit,
	}
	form.PingInfo = &models.PingInfo{
		Mode:         form.Mode,
		LatencyLimit: form.LatencyLimit,
		LostLimit:    form.LostLimit,
	}
	form.TCPInfo = &models.TCPInfo{
		EnablePorts:  form.EnablePorts,
		DisablePorts: form.DisablePorts,
	}
	return form, nil

}

func (form *RuleCreate) check() map[string]string {
	result := make(map[string]string)
	return result
}
