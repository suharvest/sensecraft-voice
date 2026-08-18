package validator

import (
	"github.com/go-playground/validator/v10"

	"github.com/YOUR-ORG/sensecraft-voice/device/agent/pkg/util"
)

func init() {
	register(
		&passwordValidator{apiValidator: newSensecraftVoiceValidator("password", "密码不符合要求，至少包含一个大写字母、一个小写字母、一个数字")},
	)
}

// passwordValidator is a customized validator for validating user password.
type passwordValidator struct {
	apiValidator
}

// validate validates the password in request.
func (pv *passwordValidator) validate(fl validator.FieldLevel) bool {
	return util.ValidateStrongPassword(fl.Field().String())
}
