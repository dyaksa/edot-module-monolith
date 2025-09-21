package usecase

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/google/uuid"
)

type warehouseTransferUsecase struct {
	db               pqsql.Database
	transferRepo     domain.WarehouseTransferRepository
	warehouseRepo    domain.WarehouseRepository
	productStockRepo domain.ProductStockRepository
	movementRepo     domain.MovementRepository
}

// CreateTransfer implements domain.WarehouseTransferUsecase.
func (wtu *warehouseTransferUsecase) CreateTransfer(ctx context.Context, req domain.CreateTransferRequest) (*domain.WarehouseTransfer, error) {
	fromWarehouseID, err := uuid.Parse(req.FromWarehouseID)
	if err != nil {
		return nil, fmt.Errorf("invalid from_warehouse_id: %w", err)
	}

	toWarehouseID, err := uuid.Parse(req.ToWarehouseID)
	if err != nil {
		return nil, fmt.Errorf("invalid to_warehouse_id: %w", err)
	}

	if fromWarehouseID == toWarehouseID {
		return nil, errors.New("cannot transfer to the same warehouse")
	}

	// Validate warehouses exist and are active
	fromWarehouse, err := wtu.warehouseRepo.Retrieve(ctx, fromWarehouseID)
	if err != nil {
		return nil, fmt.Errorf("from warehouse not found: %w", err)
	}
	if !fromWarehouse.IsActive {
		return nil, errors.New("source warehouse is not active")
	}

	toWarehouse, err := wtu.warehouseRepo.Retrieve(ctx, toWarehouseID)
	if err != nil {
		return nil, fmt.Errorf("to warehouse not found: %w", err)
	}
	if !toWarehouse.IsActive {
		return nil, errors.New("destination warehouse is not active")
	}

	// Validate warehouses belong to the same shop
	if fromWarehouse.ShopID != toWarehouse.ShopID {
		return nil, errors.New("warehouses must belong to the same shop")
	}

	var transfer *domain.WarehouseTransfer
	_, err = wtu.db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		// Create transfer
		transfer = &domain.WarehouseTransfer{
			ID:              uuid.New(),
			FromWarehouseID: fromWarehouseID,
			ToWarehouseID:   toWarehouseID,
			Status:          domain.TransferStatusRequested,
		}

		err := wtu.transferRepo.Create(ctx, tx, transfer)
		if err != nil {
			return nil, err
		}

		// Create transfer items
		var items []domain.WarehouseTransferItem
		for _, reqItem := range req.Items {
			productID, err := uuid.Parse(reqItem.ProductID)
			if err != nil {
				return nil, fmt.Errorf("invalid product_id: %w", err)
			}

			item := domain.WarehouseTransferItem{
				ID:         uuid.New(),
				TransferID: transfer.ID,
				ProductID:  productID,
				Qty:        reqItem.Qty,
			}
			items = append(items, item)
		}

		err = wtu.transferRepo.CreateItems(ctx, tx, items)
		if err != nil {
			return nil, err
		}

		transfer.Items = items
		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return transfer, nil
}

// UpdateTransferStatus implements domain.WarehouseTransferUsecase.
func (wtu *warehouseTransferUsecase) UpdateTransferStatus(ctx context.Context, transferID uuid.UUID, req domain.UpdateTransferStatusRequest) error {
	// Get current transfer
	transfer, err := wtu.transferRepo.GetByID(ctx, transferID)
	if err != nil {
		return err
	}

	// Validate status transition
	if !wtu.isValidStatusTransition(transfer.Status, req.Status) {
		return fmt.Errorf("invalid status transition from %s to %s", transfer.Status, req.Status)
	}

	// Execute the transfer if status is being set to IN_TRANSIT
	if req.Status == domain.TransferStatusInTransit {
		err = wtu.ExecuteTransfer(ctx, transferID)
		if err != nil {
			return err
		}
		return nil // ExecuteTransfer updates status internally
	}

	// For other status updates, just update the status
	_, err = wtu.db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		return nil, wtu.transferRepo.UpdateStatus(ctx, tx, transferID, req.Status)
	})

	return err
}

// GetTransfer implements domain.WarehouseTransferUsecase.
func (wtu *warehouseTransferUsecase) GetTransfer(ctx context.Context, transferID uuid.UUID) (*domain.WarehouseTransfer, error) {
	return wtu.transferRepo.GetByID(ctx, transferID)
}

// GetTransfersByWarehouse implements domain.WarehouseTransferUsecase.
func (wtu *warehouseTransferUsecase) GetTransfersByWarehouse(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]domain.WarehouseTransfer, int, error) {
	return wtu.transferRepo.GetByWarehouse(ctx, warehouseID, limit, offset)
}

// ExecuteTransfer implements domain.WarehouseTransferUsecase.
func (wtu *warehouseTransferUsecase) ExecuteTransfer(ctx context.Context, transferID uuid.UUID) error {
	// Get transfer details
	transfer, err := wtu.transferRepo.GetByID(ctx, transferID)
	if err != nil {
		return err
	}

	if transfer.Status != domain.TransferStatusRequested && transfer.Status != domain.TransferStatusApproved {
		return fmt.Errorf("cannot execute transfer with status %s", transfer.Status)
	}

	_, err = wtu.db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		// Check stock availability and reserve stock from source warehouse
		for _, item := range transfer.Items {
			// Try to reserve stock from source warehouse
			success, err := wtu.productStockRepo.TryReserveStock(ctx, tx, item.ProductID, transfer.FromWarehouseID, item.Qty)
			if err != nil {
				return nil, fmt.Errorf("failed to check stock for product %s: %w", item.ProductID, err)
			}
			if !success {
				return nil, fmt.Errorf("insufficient stock for product %s in source warehouse", item.ProductID)
			}

			// Record outbound movement from source
			err = wtu.movementRepo.Append(ctx, tx, item.ProductID, transfer.FromWarehouseID, "OUTBOUND", int(item.Qty), "TRANSFER", transfer.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to record outbound movement: %w", err)
			}

			// Commit the stock removal from source warehouse
			err = wtu.productStockRepo.CommitStock(ctx, tx, item.ProductID, transfer.FromWarehouseID, item.Qty)
			if err != nil {
				return nil, fmt.Errorf("failed to commit stock from source: %w", err)
			}
		}

		// Update transfer status to IN_TRANSIT
		err = wtu.transferRepo.UpdateStatus(ctx, tx, transferID, domain.TransferStatusInTransit)
		if err != nil {
			return nil, fmt.Errorf("failed to update transfer status: %w", err)
		}

		// Add stock to destination warehouse and mark as completed
		for _, item := range transfer.Items {
			// Record inbound movement to destination
			err = wtu.movementRepo.Append(ctx, tx, item.ProductID, transfer.ToWarehouseID, "INBOUND", int(item.Qty), "TRANSFER", transfer.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to record inbound movement: %w", err)
			}

			// Add stock to destination warehouse
			err = wtu.productStockRepo.AddStock(ctx, tx, item.ProductID, transfer.ToWarehouseID, item.Qty)
			if err != nil {
				return nil, fmt.Errorf("failed to add stock to destination: %w", err)
			}
		}

		// Mark transfer as completed
		err = wtu.transferRepo.UpdateStatus(ctx, tx, transferID, domain.TransferStatusCompleted)
		if err != nil {
			return nil, fmt.Errorf("failed to mark transfer as completed: %w", err)
		}

		return nil, nil
	})

	return err
}

// isValidStatusTransition validates if a status transition is allowed
func (wtu *warehouseTransferUsecase) isValidStatusTransition(from, to domain.TransferStatus) bool {
	switch from {
	case domain.TransferStatusRequested:
		return to == domain.TransferStatusApproved || to == domain.TransferStatusCancelled
	case domain.TransferStatusApproved:
		return to == domain.TransferStatusInTransit || to == domain.TransferStatusCancelled
	case domain.TransferStatusInTransit:
		return to == domain.TransferStatusCompleted || to == domain.TransferStatusCancelled
	case domain.TransferStatusCompleted, domain.TransferStatusCancelled:
		return false // Terminal states
	default:
		return false
	}
}

func NewWarehouseTransferUsecase(
	db pqsql.Database,
	transferRepo domain.WarehouseTransferRepository,
	warehouseRepo domain.WarehouseRepository,
	productStockRepo domain.ProductStockRepository,
	movementRepo domain.MovementRepository,
) domain.WarehouseTransferUsecase {
	return &warehouseTransferUsecase{
		db:               db,
		transferRepo:     transferRepo,
		warehouseRepo:    warehouseRepo,
		productStockRepo: productStockRepo,
		movementRepo:     movementRepo,
	}
}
