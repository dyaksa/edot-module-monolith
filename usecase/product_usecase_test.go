package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/dyaksa/warehouse/domain"
	mocks "github.com/dyaksa/warehouse/mocks/repository"
	"github.com/dyaksa/warehouse/pkg/paginator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestProductUsecase_Create_Success(t *testing.T) {
	ctx := context.Background()
	productRepo := mocks.NewMockProductRepository(t)
	stockRepo := mocks.NewMockProductStockRepository(t)

	uc := NewProductUsecase(productRepo, stockRepo)

	warehouseID := uuid.New()
	newProductID := uuid.New()

	productRepo.EXPECT().Create(ctx, mock.Anything).RunAndReturn(
		func(c context.Context, p *domain.Product) (uuid.UUID, error) {
			// Basic assertions on payload
			assert.Equal(t, "SKU-123", p.SKU)
			assert.Equal(t, "Sample", p.Name)
			return newProductID, nil
		},
	)

	stockRepo.EXPECT().Create(ctx, mock.Anything).RunAndReturn(
		func(c context.Context, ps *domain.ProductStock) (uuid.UUID, error) {
			assert.Equal(t, warehouseID, ps.WarehouseID)
			assert.Equal(t, newProductID, ps.ProductID)
			assert.Equal(t, int32(10), ps.OnHand)
			return uuid.New(), nil
		},
	)

	err := uc.Create(ctx, domain.CreateProductRequest{WarehouseID: warehouseID.String(), SKU: "SKU-123", Name: "Sample", OnHand: 10})
	assert.NoError(t, err)
}

func TestProductUsecase_Create_InvalidWarehouseUUID(t *testing.T) {
	ctx := context.Background()
	productRepo := mocks.NewMockProductRepository(t)
	stockRepo := mocks.NewMockProductStockRepository(t)
	uc := NewProductUsecase(productRepo, stockRepo)

	err := uc.Create(ctx, domain.CreateProductRequest{WarehouseID: "not-a-uuid", SKU: "SKU-123", Name: "Sample", OnHand: 1})
	assert.Error(t, err)
}

func TestProductUsecase_Create_ProductRepoError(t *testing.T) {
	ctx := context.Background()
	productRepo := mocks.NewMockProductRepository(t)
	stockRepo := mocks.NewMockProductStockRepository(t)
	uc := NewProductUsecase(productRepo, stockRepo)

	warehouseID := uuid.New()
	expectedErr := errors.New("db error")

	productRepo.EXPECT().Create(ctx, mock.Anything).RunAndReturn(
		func(c context.Context, p *domain.Product) (uuid.UUID, error) {
			return uuid.Nil, expectedErr
		},
	)

	err := uc.Create(ctx, domain.CreateProductRequest{WarehouseID: warehouseID.String(), SKU: "SKU-123", Name: "Sample", OnHand: 1})
	assert.ErrorIs(t, err, expectedErr)
}

func TestProductUsecase_RetrieveAll_Success(t *testing.T) {
	ctx := context.Background()
	productRepo := mocks.NewMockProductRepository(t)
	stockRepo := mocks.NewMockProductStockRepository(t)
	uc := NewProductUsecase(productRepo, stockRepo)

	// We expect paginator to call RetrieveAll with (limit, offset)
	products := []domain.RetrieveProduct{{SKU: "A"}, {SKU: "B"}}
	productRepo.EXPECT().RetrieveAll(ctx, 10, 0).Return(products, nil)

	req := paginator.PaginationRequest{Page: 1, Limit: 10}
	result, err := uc.RetrieveAll(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, 2, result.TotalItems)
	assert.Len(t, result.Items, 2)
}
