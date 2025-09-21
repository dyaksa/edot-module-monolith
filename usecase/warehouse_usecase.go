package usecase

import (
	"context"
	"errors"

	"github.com/dyaksa/warehouse/domain"
	"github.com/google/uuid"
)

type warehouseUsecase struct {
	warehouseRepo domain.WarehouseRepository
	transferRepo  domain.WarehouseTransferRepository
}

// Create implements domain.WarehouseUsecase.
func (w *warehouseUsecase) Create(ctx context.Context, payload domain.WarehouseCreateRequest) error {
	shopID, err := uuid.Parse(payload.ShopID)
	if err != nil {
		return err
	}

	warehouse := &domain.WareHouse{
		ShopID:   shopID,
		Name:     payload.Name,
		IsActive: payload.IsActive,
	}

	return w.warehouseRepo.Create(ctx, warehouse)
}

// Delete implements domain.WarehouseUsecase.
func (w *warehouseUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	return w.warehouseRepo.Delete(ctx, id)
}

// Retrieve implements domain.WarehouseUsecase.
func (w *warehouseUsecase) Retrieve(ctx context.Context, id uuid.UUID) (*domain.WareHouseFormatter, error) {
	warehouse, err := w.warehouseRepo.Retrieve(ctx, id)
	if err != nil {
		return nil, err
	}

	wareHouseFormatter := domain.WareHouseFormatter{
		ID:        warehouse.ID,
		Name:      warehouse.Name,
		Isactive:  warehouse.IsActive,
		CreatedAt: warehouse.CreatedAt,
	}

	return &wareHouseFormatter, nil
}

// Update implements domain.WarehouseUsecase.
func (w *warehouseUsecase) Update(ctx context.Context, id uuid.UUID, payload domain.WarehouseCreateRequest) error {
	shopID, err := uuid.Parse(payload.ShopID)
	if err != nil {
		return err
	}

	warehouse := &domain.WareHouse{
		ID:       id,
		ShopID:   shopID,
		Name:     payload.Name,
		IsActive: payload.IsActive,
	}

	return w.warehouseRepo.Update(ctx, warehouse)
}

// SetActive implements domain.WarehouseUsecase.
func (w *warehouseUsecase) SetActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	// If trying to deactivate, check for active transfers
	if !isActive {
		activeTransfers, err := w.transferRepo.GetActiveTransfersByWarehouse(ctx, id)
		if err != nil {
			return err
		}
		if len(activeTransfers) > 0 {
			return errors.New("cannot deactivate warehouse with active transfers")
		}
	}

	return w.warehouseRepo.SetActive(ctx, id, isActive)
}

// GetByShopID implements domain.WarehouseUsecase.
func (w *warehouseUsecase) GetByShopID(ctx context.Context, shopID uuid.UUID) ([]domain.WareHouseFormatter, error) {
	warehouses, err := w.warehouseRepo.GetByShopID(ctx, shopID)
	if err != nil {
		return nil, err
	}

	var formatters []domain.WareHouseFormatter
	for _, warehouse := range warehouses {
		formatter := domain.WareHouseFormatter{
			ID:        warehouse.ID,
			ShopID:    warehouse.ShopID,
			Name:      warehouse.Name,
			Isactive:  warehouse.IsActive,
			CreatedAt: warehouse.CreatedAt,
		}
		formatters = append(formatters, formatter)
	}

	return formatters, nil
}

func NewWarehouseUsecase(warehouseRepo domain.WarehouseRepository, transferRepo domain.WarehouseTransferRepository) domain.WarehouseUsecase {
	return &warehouseUsecase{
		warehouseRepo: warehouseRepo,
		transferRepo:  transferRepo,
	}
}
