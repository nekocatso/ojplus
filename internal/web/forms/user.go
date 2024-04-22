package forms

import (
	"github.com/gin-gonic/gin"
)

// User
// -Create
type UserCreate struct {
	Username     *string `validate:"required,len=12"`
	Password     *string `validate:"required,min=6,max=128"`
	Name         *string `validate:"omitempty,max=24"`
	Email        *string `validate:"required,email"`
	Verification *string `validate:"required,number,min=2,max=12" xorm:"-"`
}

func NewUserCreate(ctx *gin.Context) (*UserCreate, error) {
	var form *UserCreate
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (form *UserCreate) check() map[string]string {
	result := make(map[string]string)
	return result
}

// -Update
type UserUpdate struct {
	Email        *string `validate:"omitempty,email"`
	Verification *string `validate:"required_with=Email,number,min=2,max=12" xorm:"-"`
	Password     *string `validate:"omitempty,min=6,max=128"`
	OldPassword  *string `validate:"omitempty,required_with=Password,min=6,max=128" xorm:"-"`
	Nickname     *string `validate:"omitempty,max=12"`
}

func NewUserUpdate(ctx *gin.Context) (*UserUpdate, error) {
	var form *UserUpdate
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (form *UserUpdate) check() map[string]string {
	result := make(map[string]string)
	return result
}

// -Delete
type UserDelete struct {
	Verification *string `validate:"required,number,min=2,max=12"`
}

func NewUserDelete(ctx *gin.Context) (*UserDelete, error) {
	var form *UserDelete
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	return form, nil
}

func (form *UserDelete) check() map[string]string {
	result := make(map[string]string)
	return result
}
