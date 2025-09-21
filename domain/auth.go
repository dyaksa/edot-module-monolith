package domain

import (
	"context"
	"fmt"

	"github.com/nyaruka/phonenumbers"
)

// AuthLoginRequest represents the request payload for user authentication
type AuthLoginRequest struct {
	Identifier string `json:"identifier" validate:"required,identifier" example:"user@example.com" description:"Email address or phone number for login"`
	Password   string `json:"password" validate:"required,min=8" example:"password123" description:"User password (minimum 8 characters)"`
}

// AuthRegisterRequest represents the request payload for user registration
type AuthRegisterRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com" description:"Valid email address"`
	Phone    string `json:"phone" validate:"required,phone" example:"+6281234567890" description:"Phone number with country code"`
	Password string `json:"password" validate:"required,min=8" example:"password123" description:"Password (minimum 8 characters)"`
}

func (u AuthRegisterRequest) MustFormattedPhone() string {
	phoneNumber, err := phonenumbers.Parse(u.Phone, "ID")
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%d%d", phoneNumber.GetCountryCode(), phoneNumber.GetNationalNumber())
}

type AuthUsecase interface {
	Register(ctx context.Context, payload AuthRegisterRequest) (User, error)
	Login(ctx context.Context, payload AuthLoginRequest) (string, error)
}
