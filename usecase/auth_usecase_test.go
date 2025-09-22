package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/dyaksa/encryption-pii/crypto/aesx"
	"github.com/dyaksa/encryption-pii/crypto/core"
	"github.com/dyaksa/warehouse/bootstrap"
	"github.com/dyaksa/warehouse/domain"
	mocks "github.com/dyaksa/warehouse/mocks/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// simpleCryptoStub implements infraCrypto.Crypto
type simpleCryptoStub struct{}

func (s simpleCryptoStub) AESFunc() func() (core.PrimitiveAES, error) {
	return func() (core.PrimitiveAES, error) { var p core.PrimitiveAES; return p, nil }
}
func (s simpleCryptoStub) Encrypt(data string) aesx.AES[string, core.PrimitiveAES] {
	return aesx.AESChiper(s.AESFunc(), data, aesx.AesCBC)
}
func (s simpleCryptoStub) Decrypt(def string) aesx.AES[string, core.PrimitiveAES] {
	return aesx.AESChiper(s.AESFunc(), def, aesx.AesCBC)
}
func (s simpleCryptoStub) BindHeap(entity any) error  { return nil }
func (s simpleCryptoStub) HashString(v string) string { return "hash(" + v + ")" }

func TestAuthUsecase_Login_InvalidIdentifier(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewMockUserRepository(t)
	c := simpleCryptoStub{}
	env := &bootstrap.Env{JwtSecret: "secret", JwtExpiry: 3600}
	uc := NewAuthUsecase(userRepo, c, env)

	token, err := uc.Login(ctx, domain.AuthLoginRequest{Identifier: "!!!", Password: "pwd123456"})
	assert.Error(t, err)
	assert.Empty(t, token)
}

func TestAuthUsecase_Register_Success(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewMockUserRepository(t)
	c := simpleCryptoStub{}
	env := &bootstrap.Env{JwtSecret: "secret", JwtExpiry: 3600}
	uc := NewAuthUsecase(userRepo, c, env)

	reg := domain.AuthRegisterRequest{Email: "user@example.com", Phone: "+6281234567890", Password: "password123"}

	userRepo.EXPECT().CreateUser(ctx, mock.Anything).Return(nil)

	user, err := uc.Register(ctx, reg)
	assert.NoError(t, err)
	// cannot directly assert internal encrypted value; ensure struct not empty
	assert.NotEmpty(t, user.Email)
}

func TestAuthUsecase_Register_CreateError(t *testing.T) {
	ctx := context.Background()
	userRepo := mocks.NewMockUserRepository(t)
	c := simpleCryptoStub{}
	env := &bootstrap.Env{JwtSecret: "secret", JwtExpiry: 3600}
	uc := NewAuthUsecase(userRepo, c, env)

	reg := domain.AuthRegisterRequest{Email: "user@example.com", Phone: "+6281234567890", Password: "password123"}
	expectedErr := errors.New("insert fail")
	userRepo.EXPECT().CreateUser(ctx, mock.Anything).Return(expectedErr)

	_, err := uc.Register(ctx, reg)
	assert.ErrorIs(t, err, expectedErr)
}
