package forms

import (
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zhTrans "github.com/go-playground/validator/v10/translations/zh"
)

type Form interface {
	check() map[string]string
}

func Verify(form Form) (bool, map[string]map[string]string, error) {
	validate := validator.New()
	uniTrans := ut.New(zh.New())
	trans, _ := uniTrans.GetTranslator("zh")
	zhTrans.RegisterDefaultTranslations(validate, trans)
	errorMap := make(map[string]string)
	flag := true
	err := validate.Struct(form)
	if err != nil {
		if errV, ok := err.(validator.ValidationErrors); ok {
			flag = false
			for _, err := range errV {
				errorMap[err.Field()] = err.Translate(trans)
			}
		} else {
			return false, nil, err
		}
	}
	cleanMap := form.check()
	for key, value := range cleanMap {
		_, ok := errorMap[key]
		if !ok {
			errorMap[key] = value
		}
	}
	result := map[string]map[string]string{"errors": errorMap}
	return flag, result, nil
}
