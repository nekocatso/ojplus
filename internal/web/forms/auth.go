package forms

import (
	"github.com/gin-gonic/gin"
)

type TokenCreate struct {
	Account      *string `validate:"omitempty,max=32"`
	Password     *string `validate:"omitempty,required_with=Account"`
	Email        *string `validate:"omitempty,email"`
	Verification *string `validate:"omitempty,required_with=Email,number,min=2,max=12" xorm:"-"`
}

func NewTokenCreate(ctx *gin.Context) (*TokenCreate, error) {
	var form *TokenCreate
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (form *TokenCreate) check() map[string]string {
	result := make(map[string]string)
	return result
}

type TokenRefresh struct {
	RefreshToken string `validate:"required"`
}

func NewTokenRefresh(ctx *gin.Context) (*TokenRefresh, error) {
	var form *TokenRefresh
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (form *TokenRefresh) check() map[string]string {
	result := make(map[string]string)
	return result
}

type EmailRequire struct {
	Email *string `validate:"required,email"`
}

func NewEmailRequire(ctx *gin.Context) (*EmailRequire, error) {
	var form *EmailRequire
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (form *EmailRequire) check() map[string]string {
	result := make(map[string]string)
	return result
}

type EmailOmitempty struct {
	Email *string `validate:"omitempty,email"`
}

func NewEmailOmitempty(ctx *gin.Context) (*EmailOmitempty, error) {
	var form *EmailOmitempty
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (form *EmailOmitempty) check() map[string]string {
	result := make(map[string]string)
	return result
}
