package forms

import "github.com/gin-gonic/gin"

type RulePingCreate struct {
	Name         string `validate:"required,max=24"`
	Type         string `validate:"required,max=12"`
	AlarmID      int    `validate:"required"`
	Overtime     int    `validate:"required"`
	DeclineLimit int    `validate:"required"`
	RecoverLimit int    `validate:"required"`
	LatencyLimit int    `validate:"required"`
	LostLimit    int    `validate:"required"`
	Mode         int    `validate:"required"`
}

func NewRuleCreate(ctx *gin.Context) (*RulePingCreate, error) {
	var form *RulePingCreate
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (form *RulePingCreate) check() map[string]string {
	result := make(map[string]string)
	return result
}
