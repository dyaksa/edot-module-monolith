package helper

import (
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/nyaruka/phonenumbers"
)

func NormalizeIdentifier(raw string) (kind string, normalized string, ok bool) {
	s := strings.TrimSpace(raw)

	if validator.New().Var(s, "email") == nil {
		parts := strings.Split(s, "@")
		if len(parts) == 2 {
			return "email", parts[0] + "@" + strings.ToLower(parts[1]), true
		}
	}

	if num, err := phonenumbers.Parse(s, "ID"); err == nil && phonenumbers.IsValidNumber(num) {
		return "phone", phonenumbers.Format(num, phonenumbers.E164), true
	}

	return "", "", false
}
