package validationutils

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/nyaruka/phonenumbers"
)

func IdentifierValidator(v *validator.Validate) error {
	return v.RegisterValidation("identifier", func(fl validator.FieldLevel) bool {
		raw := strings.TrimSpace(fl.Field().String())
		if raw == "" {
			return false
		}
		// Cek email sederhana (re-usable validator bawaan)
		if err := validator.New().Var(raw, "email"); err == nil {
			return true
		}
		// Cek phone via libphonenumber (anggap default region "ID")
		num, err := phonenumbers.Parse(raw, "ID")
		return err == nil && phonenumbers.IsValidNumber(num)
	})
}
