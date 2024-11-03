// utils/validator.go
package utils

import (
	"github.com/go-playground/validator/v10"
)

// Validator adalah instance global dari validator
var Validator *validator.Validate

// InitValidator menginisialisasi validator
func InitValidator() {
    Validator = validator.New()
}
