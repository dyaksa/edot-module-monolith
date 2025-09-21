package usecase

import (
	"context"
	"errors"

	"github.com/dyaksa/warehouse/bootstrap"
	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/crypto"
	"github.com/dyaksa/warehouse/pkg/errx"
	"github.com/dyaksa/warehouse/pkg/helper"
	"github.com/dyaksa/warehouse/pkg/passwordutils"
	"github.com/dyaksa/warehouse/pkg/tokenutils"
)

type authUsecase struct {
	userRepo domain.UserRepository
	crypto   crypto.Crypto
	env      *bootstrap.Env
}

// Login implements domain.AuthUsecase.
func (a *authUsecase) Login(ctx context.Context, payload domain.AuthLoginRequest) (string, error) {
	_, norm, ok := helper.NormalizeIdentifier(payload.Identifier)
	if !ok {
		return "", errx.E(errx.CodeValidation, "invalid identifier", errx.Op("authUsecase.Login"))
	}

	existsUser, err := a.userRepo.GetMailOrPhone(ctx, a.crypto.HashString(norm), a.crypto.HashString(norm), func(data *domain.User) {
		data.Email = a.crypto.Decrypt("")
		data.Phone = a.crypto.Decrypt("")
	})

	if ok := passwordutils.VerifyPassword(payload.Password, existsUser.PasswordHash); !ok {
		return "", errors.New("invalid credentials")
	}

	if err != nil {
		return "", errx.E(errx.CodeNotFound, "user not found", errx.Op("authUsecase.Login"), err)
	}

	if ok := passwordutils.VerifyPassword(payload.Password, existsUser.PasswordHash); !ok {
		return "", errx.E(errx.CodeUnauthorized, "invalid credentials", errx.Op("authUsecase.Login"), errors.New("password mismatch"))
	}

	accessToken, err := tokenutils.CreateAccessToken(existsUser, a.env.JwtSecret, a.env.JwtExpiry)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// Register implements domain.AuthUsecase.
func (a *authUsecase) Register(ctx context.Context, payload domain.AuthRegisterRequest) (domain.User, error) {
	var user domain.User
	user.Email = a.crypto.Encrypt(payload.Email)
	user.Phone = a.crypto.Encrypt(payload.MustFormattedPhone())

	passwordHash, err := passwordutils.HashPassword(payload.Password)
	if err != nil {
		return user, errx.E(errx.CodeInternal, "failed to hash password", errx.Op("authUsecase.Register"), err)
	}

	user.PasswordHash = passwordHash
	if err := a.crypto.BindHeap(&user); err != nil {
		return user, errx.E(errx.CodeInternal, "failed to bind crypto", errx.Op("authUsecase.Register"), err)
	}

	user.PhoneBidx = a.crypto.HashString(payload.MustFormattedPhone())
	err = a.userRepo.CreateUser(ctx, &user)
	if err != nil {
		return user, errx.E(errx.CodeInternal, "failed to create user", errx.Op("authUsecase.Register"), err)
	}

	return user, nil
}

func NewAuthUsecase(
	userRepo domain.UserRepository,
	crypto crypto.Crypto,
	env *bootstrap.Env,
) domain.AuthUsecase {
	return &authUsecase{
		userRepo: userRepo,
		crypto:   crypto,
		env:      env,
	}
}
