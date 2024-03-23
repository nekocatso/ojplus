package forms

import (
	"Alarm/internal/web/models"
	"regexp"

	"github.com/gin-gonic/gin"
)

type UserCreate struct {
	Username string `Name:"username" validate:"required"`
	Password string `Name:"password" validate:"required"`
	Name     string `Name:"name" validate:"required"`
	Email    string `Name:"email" validate:"omitempty,email"`
	Phone    string `Name:"phone" validate:"omitempty,number"`
	Model    *models.User
}

func NewUserCreate(ctx *gin.Context) (*UserCreate, error) {
	var form *UserCreate
	err := ctx.BindJSON(&form)
	if err != nil {
		return nil, err
	}
	form.Model = &models.User{
		Username: form.Username,
		Password: form.Password,
		Name:     form.Name,
		Email:    form.Email,
		Phone:    form.Phone,
	}
	return form, nil
}

func (form *UserCreate) check() map[string]string {
	result := make(map[string]string)
	if form.Phone != "" && !checkPhone(form.Phone) {
		result["phone"] = "电话号码格式错误"
	}
	return result
}

func checkPhone(phone string) bool {
	regex := `^(?:\+)?[0-9]{1,3}[-.●]?\(?[0-9]{1,3}\)?[-.●]?[0-9]{1,4}[-.●]?[0-9]{1,4}$`
	pattern := regexp.MustCompile(regex)
	matched := pattern.MatchString(phone)
	return matched
}
