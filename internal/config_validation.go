package internal

import (
	"github.com/go-playground/validator/v10"
	"sync"
)

var (
	once     sync.Once
	validate *validator.Validate
)

func ValidateConfig(config *ConditionalRebootConfig) error {
	once.Do(func() {
		validate = validator.New()
	})

	return validate.Struct(config)
}
