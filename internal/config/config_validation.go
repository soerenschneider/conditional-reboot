package config

import (
	"github.com/go-playground/validator/v10"
)

var (
	validate = validator.New()
)

func Validate(config *ConditionalRebootConfig) error {
	return validate.Struct(config)
}
