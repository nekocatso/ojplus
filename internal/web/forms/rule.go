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
	Info         *typeInfo

	Model    *models.Rule     `validate:"-"`
	PingInfo *models.PingInfo `validate:"-"`
	TCPInfo  *models.TCPInfo  `validate:"-"`
}

type typeInfo struct {
	Mode         int    `validate:"required_with=LatencyLimit LostLimit"`
	LatencyLimit int    `validate:"required_with=Mode LostLimit"`
	LostLimit    int    `validate:"required_with=Mode LatencyLimit"`
	EnablePorts  string `validate:"required_with=DisablePorts,max=128"`
	DisablePorts string `validate:"required_with=EnablePorts,max=128"`
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
	return result
}
