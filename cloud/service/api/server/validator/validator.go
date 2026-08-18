package validator

import (
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zt "github.com/go-playground/validator/v10/translations/zh"
)

type customValidator interface {
	getTag() string
	translateError(ut ut.Translator) error
	translate(ut ut.Translator, fe validator.FieldError) string

	// Should be implemented by the custom validator.
	validate(fl validator.FieldLevel) bool
}

var tran ut.Translator
var customValidators []customValidator

// register adds a new custom validator to the validator list
func register(validators ...customValidator) {
	customValidators = append(customValidators, validators...)
}

func init() {
	_zh := zh.New() // default is Chinese
	uni := ut.New(_zh, _zh)
	tran, _ = uni.GetTranslator("zh")

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_ = zt.RegisterDefaultTranslations(v, tran)

		for _, c := range customValidators {
			_ = v.RegisterValidation(c.getTag(), c.validate)
			_ = v.RegisterTranslation(c.getTag(), tran, c.translateError, c.translate)
		}
	}
}

type apiValidator struct {
	tag string
	err string
}

func newSensecraftVoiceValidator(tag, err string) apiValidator {
	return apiValidator{
		tag: tag,
		err: err,
	}
}

func (c apiValidator) getTag() string {
	return c.tag
}

func (c apiValidator) translateError(ut ut.Translator) error {
	return ut.Add(c.tag, "{0}"+c.err, true)
}

func (c apiValidator) translate(ut ut.Translator, fe validator.FieldError) string {
	t, _ := ut.T(c.tag, fe.Field())
	return t
}
