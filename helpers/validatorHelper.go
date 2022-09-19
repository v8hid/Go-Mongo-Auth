package helpers

import (
	"fmt"
	"strings"
	"v8hid/Go-Mongo-Auth/models"

	"github.com/go-playground/validator/v10"
)

func MakeValidationErrors(rawErrors validator.ValidationErrors) []models.ValidationError {
	finalErrors := make([]models.ValidationError, len(rawErrors))
	for k, err := range rawErrors {
		fe := models.ValidationError{
			Field: strings.ToLower(err.Field()),
			Msg:   msgForTag(err),
		}
		finalErrors[k] = fe
	}
	return finalErrors
}
func msgForTag(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "Is required"
	case "email":
		return "Invalid email"
	case "min":
		return fmt.Sprintf("Should have at least %v charachters", err.Param())
	case "max":
		return fmt.Sprintf("Should be less than %v charachters", err.Param())
	case "eq=USER|eq=ADMIN":
		return "Should be USER or ADMIN"
	}
	return "Wrong input"
}
