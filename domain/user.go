package domain

import (
	"context"

	"github.com/dyaksa/encryption-pii/crypto/types"
	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID       `json:"id"`
	Email        types.AESCipher `json:"email"`
	Phone        types.AESCipher `json:"phone"`
	EmailBidx    string          `json:"email_bidx" full_text_search:"true"`
	PhoneBidx    string          `json:"phone_bidx" full_text_search:"true"`
	PasswordHash string          `json:"password_hash"`
}

//go:generate mockery
type UserRepository interface {
	CreateUser(ctx context.Context, user *User) error
	GetUserByEmail(ctx context.Context, email_bidx string, fn func(data *User)) (*User, error)
	GetUserByPhone(ctx context.Context, phone_bidx string, fn func(data *User)) (*User, error)
	GetMailOrPhone(ctx context.Context, email_bidx, phone_bidx string, fn func(data *User)) (*User, error)
}
