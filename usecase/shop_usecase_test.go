package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/dyaksa/warehouse/domain"
	mocks "github.com/dyaksa/warehouse/mocks/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestShopUsecase_Create_Success(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockShopRepository(t)
	uc := NewShopUsecase(repo)

	repo.EXPECT().Create(ctx, mock.Anything).RunAndReturn(
		func(c context.Context, s *domain.Shop) (uuid.UUID, error) {
			assert.Equal(t, "Main Shop", s.Name)
			return uuid.New(), nil
		},
	)

	err := uc.Create(ctx, domain.CreateShopRequest{Name: "Main Shop"})
	assert.NoError(t, err)
}

func TestShopUsecase_Create_RepoError(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockShopRepository(t)
	uc := NewShopUsecase(repo)
	expected := errors.New("insert failed")

	repo.EXPECT().Create(ctx, mock.Anything).Return(uuid.Nil, expected)

	err := uc.Create(ctx, domain.CreateShopRequest{Name: "X"})
	assert.ErrorIs(t, err, expected)
}

func TestShopUsecase_Retrieve_Success(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockShopRepository(t)
	uc := NewShopUsecase(repo)
	id := uuid.New()
	repo.EXPECT().Retrieve(ctx, id).Return(&domain.Shop{ID: id, Name: "Shop1"}, nil)

	shop, err := uc.Retrieve(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, id, shop.ID)
}

func TestShopUsecase_Retrieve_Error(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockShopRepository(t)
	uc := NewShopUsecase(repo)
	id := uuid.New()
	expected := errors.New("not found")
	repo.EXPECT().Retrieve(ctx, id).Return(nil, expected)

	shop, err := uc.Retrieve(ctx, id)
	assert.ErrorIs(t, err, expected)
	assert.Nil(t, shop)
}

func TestShopUsecase_Update_Success(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockShopRepository(t)
	uc := NewShopUsecase(repo)
	id := uuid.New()
	repo.EXPECT().Update(ctx, mock.Anything).RunAndReturn(
		func(c context.Context, s *domain.Shop) error {
			assert.Equal(t, id, s.ID)
			assert.Equal(t, "Updated", s.Name)
			return nil
		},
	)
	err := uc.Update(ctx, domain.UpdateShopRequest{ID: id, Name: "Updated"})
	assert.NoError(t, err)
}

func TestShopUsecase_Update_Error(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockShopRepository(t)
	uc := NewShopUsecase(repo)
	id := uuid.New()
	expected := errors.New("update failed")
	repo.EXPECT().Update(ctx, mock.Anything).Return(expected)
	err := uc.Update(ctx, domain.UpdateShopRequest{ID: id, Name: "X"})
	assert.ErrorIs(t, err, expected)
}

func TestShopUsecase_Delete_Success(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockShopRepository(t)
	uc := NewShopUsecase(repo)
	id := uuid.New()
	repo.EXPECT().Delete(ctx, id).Return(nil)
	err := uc.Delete(ctx, id)
	assert.NoError(t, err)
}

func TestShopUsecase_Delete_Error(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockShopRepository(t)
	uc := NewShopUsecase(repo)
	id := uuid.New()
	expected := errors.New("delete failed")
	repo.EXPECT().Delete(ctx, id).Return(expected)
	err := uc.Delete(ctx, id)
	assert.ErrorIs(t, err, expected)
}

