package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/dyaksa/warehouse/domain"
	mocks "github.com/dyaksa/warehouse/mocks/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWarehouseUsecase_Create_Success(t *testing.T) {
	ctx := context.Background()
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	transferRepo := mocks.NewMockWarehouseTransferRepository(t)
	uc := NewWarehouseUsecase(warehouseRepo, transferRepo)

	shopID := uuid.New()
	warehouseRepo.EXPECT().Create(ctx, mock.Anything).Return(nil).Run(func(ctx context.Context, w *domain.WareHouse) {
		assert.Equal(t, shopID, w.ShopID)
		assert.Equal(t, "Main", w.Name)
		assert.True(t, w.IsActive)
	})

	err := uc.Create(ctx, domain.WarehouseCreateRequest{ShopID: shopID.String(), Name: "Main", IsActive: true})
	assert.NoError(t, err)
}

func TestWarehouseUsecase_Create_InvalidShopID(t *testing.T) {
	ctx := context.Background()
	uc := NewWarehouseUsecase(nil, nil)
	err := uc.Create(ctx, domain.WarehouseCreateRequest{ShopID: "bad", Name: "Main", IsActive: true})
	assert.Error(t, err)
}

func TestWarehouseUsecase_Retrieve_Success(t *testing.T) {
	ctx := context.Background()
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	transferRepo := mocks.NewMockWarehouseTransferRepository(t)
	uc := NewWarehouseUsecase(warehouseRepo, transferRepo)

	id := uuid.New()
	wh := &domain.WareHouse{ID: id, Name: "Main", IsActive: true, CreatedAt: time.Now()}
	warehouseRepo.EXPECT().Retrieve(ctx, id).Return(wh, nil)

	res, err := uc.Retrieve(ctx, id)
	assert.NoError(t, err)
	assert.Equal(t, wh.ID, res.ID)
	assert.Equal(t, wh.Name, res.Name)
	assert.Equal(t, wh.IsActive, res.Isactive)
}

func TestWarehouseUsecase_Retrieve_Error(t *testing.T) {
	ctx := context.Background()
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	uc := NewWarehouseUsecase(warehouseRepo, nil)
	id := uuid.New()
	warehouseRepo.EXPECT().Retrieve(ctx, id).Return(nil, errors.New("not found"))
	res, err := uc.Retrieve(ctx, id)
	assert.Error(t, err)
	assert.Nil(t, res)
}

func TestWarehouseUsecase_Update_Success(t *testing.T) {
	ctx := context.Background()
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	uc := NewWarehouseUsecase(warehouseRepo, nil)
	id := uuid.New()
	shopID := uuid.New()
	warehouseRepo.EXPECT().Update(ctx, mock.Anything).Return(nil).Run(func(ctx context.Context, w *domain.WareHouse) {
		assert.Equal(t, id, w.ID)
		assert.Equal(t, shopID, w.ShopID)
		assert.Equal(t, "Updated", w.Name)
		assert.False(t, w.IsActive)
	})
	err := uc.Update(ctx, id, domain.WarehouseCreateRequest{ShopID: shopID.String(), Name: "Updated", IsActive: false})
	assert.NoError(t, err)
}

func TestWarehouseUsecase_Update_InvalidShopID(t *testing.T) {
	ctx := context.Background()
	uc := NewWarehouseUsecase(nil, nil)
	err := uc.Update(ctx, uuid.New(), domain.WarehouseCreateRequest{ShopID: "xx", Name: "Updated", IsActive: false})
	assert.Error(t, err)
}

func TestWarehouseUsecase_Delete_Success(t *testing.T) {
	ctx := context.Background()
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	uc := NewWarehouseUsecase(warehouseRepo, nil)
	id := uuid.New()
	warehouseRepo.EXPECT().Delete(ctx, id).Return(nil)
	err := uc.Delete(ctx, id)
	assert.NoError(t, err)
}

func TestWarehouseUsecase_SetActive_DeactivateWithActiveTransfers_Error(t *testing.T) {
	ctx := context.Background()
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	transferRepo := mocks.NewMockWarehouseTransferRepository(t)
	uc := NewWarehouseUsecase(warehouseRepo, transferRepo)
	id := uuid.New()
	transferRepo.EXPECT().GetActiveTransfersByWarehouse(ctx, id).Return([]domain.WarehouseTransfer{{ID: uuid.New()}}, nil)
	err := uc.SetActive(ctx, id, false)
	assert.Error(t, err)
}

func TestWarehouseUsecase_SetActive_Deactivate_NoActiveTransfers_Success(t *testing.T) {
	ctx := context.Background()
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	transferRepo := mocks.NewMockWarehouseTransferRepository(t)
	uc := NewWarehouseUsecase(warehouseRepo, transferRepo)
	id := uuid.New()
	transferRepo.EXPECT().GetActiveTransfersByWarehouse(ctx, id).Return([]domain.WarehouseTransfer{}, nil)
	warehouseRepo.EXPECT().SetActive(ctx, id, false).Return(nil)
	err := uc.SetActive(ctx, id, false)
	assert.NoError(t, err)
}

func TestWarehouseUsecase_SetActive_Activate_Success(t *testing.T) {
	ctx := context.Background()
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	uc := NewWarehouseUsecase(warehouseRepo, nil)
	id := uuid.New()
	warehouseRepo.EXPECT().SetActive(ctx, id, true).Return(nil)
	err := uc.SetActive(ctx, id, true)
	assert.NoError(t, err)
}

func TestWarehouseUsecase_GetByShopID_Success(t *testing.T) {
	ctx := context.Background()
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	uc := NewWarehouseUsecase(warehouseRepo, nil)
	shopID := uuid.New()
	warehouses := []domain.WareHouse{{ID: uuid.New(), ShopID: shopID, Name: "A", IsActive: true}, {ID: uuid.New(), ShopID: shopID, Name: "B", IsActive: false}}
	warehouseRepo.EXPECT().GetByShopID(ctx, shopID).Return(warehouses, nil)
	res, err := uc.GetByShopID(ctx, shopID)
	assert.NoError(t, err)
	assert.Len(t, res, 2)
	assert.Equal(t, warehouses[0].Name, res[0].Name)
	assert.Equal(t, warehouses[1].IsActive, res[1].Isactive)
}

func TestWarehouseUsecase_GetByShopID_Error(t *testing.T) {
	ctx := context.Background()
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	uc := NewWarehouseUsecase(warehouseRepo, nil)
	shopID := uuid.New()
	warehouseRepo.EXPECT().GetByShopID(ctx, shopID).Return(nil, errors.New("db err"))
	res, err := uc.GetByShopID(ctx, shopID)
	assert.Error(t, err)
	assert.Nil(t, res)
}
