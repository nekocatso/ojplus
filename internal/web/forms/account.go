package forms

import (
	"Alarm/internal/web/models"
	"regexp"

	"github.com/gin-gonic/gin"
)

type UserCreate struct {
	Username string       `validate:"required,min=3,max=32"`
	Password string       `validate:"required,min=6,max=128"`
	Name     string       `validate:"required,max=24"`
	Email    string       `validate:"omitempty,email"`
	Phone    string       `validate:"required,number,min=6,max=24"`
	Role     int          `validate:"required,oneof=10 20 30"`
	Model    *models.User `validate:"-"`
}

func NewUserCreate(ctx *gin.Context) (*UserCreate, error) {
	var form *UserCreate
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	form.Model = &models.User{
		Username: form.Username,
		Password: form.Password,
		Name:     form.Name,
		Email:    form.Email,
		Phone:    form.Phone,
		Role:     form.Role,
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

type UserUpdate struct {
	Email       string                 `validate:"omitempty,email"`
	Phone       string                 `validate:"omitempty,number,min=6,max=32"`
	Note        string                 `validate:"omitempty,max=128"`
	OldPassword string                 `validate:"omitempty,required_with=Password,min=6,max=128"`
	Password    string                 `validate:"omitempty,min=6,max=128"`
	IsResetPwd  bool                   `validate:"omitempty"`
	UpdateMap   map[string]interface{} `validate:"-"`
}

func NewUserUpdate(ctx *gin.Context) (*UserUpdate, error) {
	var form *UserUpdate
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	form.UpdateMap = map[string]interface{}{}
	if form.Email != "" {
		form.UpdateMap["email"] = form.Email
	}
	if form.Phone != "" {
		form.UpdateMap["phone"] = form.Phone
	}
	if form.Note != "" {
		form.UpdateMap["note"] = form.Note
	}
	if form.Password != "" {
		form.UpdateMap["password"] = form.Password
	}
	return form, nil
}

func (form *UserUpdate) check() map[string]string {
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

type UserSelect struct {
	Page       int `validate:"required,gt=0"`
	PageSize   int `validate:"required,gt=0,lte=100"`
	Query      *UserConditions
	Model      *models.User           `validate:"-"`
	Conditions map[string]interface{} `validate:"-"`
}

type UserConditions struct {
	Username        string `validate:"omitempty"`
	Name            string `validate:"omitempty"`
	Phone           string `validate:"omitempty"`
	IsActive        int    `validate:"omitempty,oneof=-1 1"`
	Role            int    `validate:"omitempty,oneof=10 20 30"`
	CreateTimeBegin int    `validate:"required_with=CreateTimeEnd,gte=0"`
	CreateTimeEnd   int    `validate:"required_with=CreateTimeBegin,gtefield=CreateTimeBegin"`
}

func NewUserSelect(ctx *gin.Context) (*UserSelect, error) {
	var form *UserSelect
	err := ctx.ShouldBind(&form)
	if err != nil {
		return nil, err
	}
	if form.Query == nil {
		form.Query = &UserConditions{}
	}
	form.Model = &models.User{
		Username: form.Query.Username,
		Name:     form.Query.Name,
	}
	form.Conditions = make(map[string]interface{})
	if form.Query.Username != "" {
		form.Conditions["username"] = form.Query.Username
	}
	if form.Query.Name != "" {
		form.Conditions["name"] = form.Query.Name
	}
	if form.Query.Phone != "" {
		form.Conditions["phone"] = form.Query.Phone
	}
	if form.Query.IsActive == -1 {
		form.Conditions["isActive"] = 0
	} else if form.Query.IsActive == 1 {
		form.Conditions["isActive"] = 1
	}
	if form.Query.Role != 0 {
		form.Conditions["role"] = form.Query.Role
	}
	if form.Query.CreateTimeBegin != 0 {
		form.Conditions["createTimeBegin"] = form.Query.CreateTimeBegin
	}
	if form.Query.CreateTimeEnd != 0 {
		form.Conditions["createTimeEnd"] = form.Query.CreateTimeEnd
	}
	return form, nil
}

func (form *UserSelect) check() map[string]string {
	result := make(map[string]string)
	return result
}
