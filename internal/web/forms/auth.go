package forms

import (
	"Alarm/internal/web/models"

	"github.com/gin-gonic/gin"
)

type Login struct {
	Username string       `validate:"required,max=32"`
	Password string       `validate:"required,max=32"`
	Model    *models.User `validate:"-"`
}

func NewLogin(ctx *gin.Context) (*Login, error) {
	var form *Login
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	form.Model = &models.User{
		Username: form.Username,
		Password: form.Password,
	}
	return form, nil
}

func (form *Login) check() map[string]string {
	result := make(map[string]string)
	return result
}

type Refresh struct {
	RefreshToken string `validate:"required"`
}

func NewRefresh(ctx *gin.Context) (*Refresh, error) {
	var form *Refresh
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (form *Refresh) check() map[string]string {
	result := make(map[string]string)
	return result
}
