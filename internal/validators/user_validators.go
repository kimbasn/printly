package validators

import (
	"strings"

	"gopkg.in/go-playground/validator.v9"
)

func ValidateRole(fl validator.FieldLevel) bool {
	role := strings.ToLower(fl.Field().String())
	return role == "user" || role == "manager" || role == "admin"
}
