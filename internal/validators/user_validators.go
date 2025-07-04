package validators

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

func ValidateRole(fl validator.FieldLevel) bool {
	role := strings.ToLower(fl.Field().String())
	return role == "user" || role == "manager" || role == "admin"
}
