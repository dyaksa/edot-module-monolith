package domain

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// Transfer status enum
type TransferStatus string

const (
	TransferStatusRequested TransferStatus = "REQUESTED"
	TransferStatusApproved  TransferStatus = "APPROVED"
	TransferStatusInTransit TransferStatus = "IN_TRANSIT"
	TransferStatusCompleted TransferStatus = "COMPLETED"
	TransferStatusCancelled TransferStatus = "CANCELLED"
)

type WareHouse struct {
	ID        uuid.UUID
	ShopID    uuid.UUID
	Name      string
	IsActive  bool
	CreatedAt time.Time
}

// WarehouseTransfer represents a transfer between warehouses
type WarehouseTransfer struct {
	ID              uuid.UUID               `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" description:"Transfer UUID"`
	FromWarehouseID uuid.UUID               `json:"from_warehouse_id" example:"550e8400-e29b-41d4-a716-446655440001" description:"Source warehouse UUID"`
	ToWarehouseID   uuid.UUID               `json:"to_warehouse_id" example:"550e8400-e29b-41d4-a716-446655440002" description:"Destination warehouse UUID"`
	Status          TransferStatus          `json:"status" example:"REQUESTED" description:"Transfer status: REQUESTED, APPROVED, IN_TRANSIT, COMPLETED, CANCELLED"`
	CreatedAt       time.Time               `json:"created_at" example:"2024-01-15T10:30:00Z" description:"Transfer creation timestamp"`
	Items           []WarehouseTransferItem `json:"items,omitempty" description:"List of items being transferred"`
}

// WarehouseTransferItem represents a product item in a transfer
type WarehouseTransferItem struct {
	ID         uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440003" description:"Transfer item UUID"`
	TransferID uuid.UUID `json:"transfer_id" example:"550e8400-e29b-41d4-a716-446655440000" description:"Parent transfer UUID"`
	ProductID  uuid.UUID `json:"product_id" example:"550e8400-e29b-41d4-a716-446655440004" description:"Product UUID being transferred"`
	Qty        int32     `json:"qty" example:"10" description:"Quantity to transfer"`
}

// WareHouseFormatter represents the response format for warehouse data
type WareHouseFormatter struct {
	ID        uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" description:"Warehouse UUID"`
	ShopID    uuid.UUID `json:"shop_id" example:"550e8400-e29b-41d4-a716-446655440001" description:"Shop UUID that owns this warehouse"`
	Name      string    `json:"name" example:"Main Warehouse" description:"Warehouse name"`
	Isactive  bool      `json:"is_active" example:"true" description:"Whether the warehouse is active"`
	CreatedAt time.Time `json:"created_at" example:"2024-01-15T10:30:00Z" description:"Warehouse creation timestamp"`
}

// WarehouseCreateRequest represents the request payload for creating/updating warehouses
type WarehouseCreateRequest struct {
	ShopID   string `json:"shop_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000" description:"UUID of the shop that owns this warehouse"`
	Name     string `json:"name" binding:"required" example:"Main Warehouse" description:"Warehouse name"`
	IsActive bool   `json:"is_active" example:"true" description:"Whether the warehouse should be active"`
}

// CreateTransferRequest represents the request payload for creating warehouse transfers
type CreateTransferRequest struct {
	FromWarehouseID string                      `json:"from_warehouse_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440001" description:"Source warehouse UUID"`
	ToWarehouseID   string                      `json:"to_warehouse_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440002" description:"Destination warehouse UUID"`
	Items           []CreateTransferItemRequest `json:"items" binding:"required,dive" description:"List of items to transfer"`
}

// CreateTransferItemRequest represents an item in a transfer request
type CreateTransferItemRequest struct {
	ProductID string `json:"product_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440004" description:"Product UUID to transfer"`
	Qty       int32  `json:"qty" binding:"required,min=1" example:"10" description:"Quantity to transfer (must be positive)"`
}

// UpdateTransferStatusRequest represents the request payload for updating transfer status
type UpdateTransferStatusRequest struct {
	Status TransferStatus `json:"status" binding:"required" example:"APPROVED" description:"New transfer status: REQUESTED, APPROVED, IN_TRANSIT, COMPLETED, CANCELLED"`
}

type WarehouseRepository interface {
	Pick(ctx context.Context, tx *sql.Tx, productID uuid.UUID, qty int, shopID uuid.UUID) (warehouseID uuid.UUID, err error)
	Create(ctx context.Context, w *WareHouse) error
	Retrieve(ctx context.Context, id uuid.UUID) (*WareHouse, error)
	Update(ctx context.Context, w *WareHouse) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByShopID(ctx context.Context, shopID uuid.UUID) ([]WareHouse, error)
	SetActive(ctx context.Context, id uuid.UUID, isActive bool) error
}

type WarehouseTransferRepository interface {
	Create(ctx context.Context, tx *sql.Tx, transfer *WarehouseTransfer) error
	GetByID(ctx context.Context, id uuid.UUID) (*WarehouseTransfer, error)
	UpdateStatus(ctx context.Context, tx *sql.Tx, id uuid.UUID, status TransferStatus) error
	GetByWarehouse(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]WarehouseTransfer, int, error)
	GetActiveTransfersByWarehouse(ctx context.Context, warehouseID uuid.UUID) ([]WarehouseTransfer, error)
	CreateItems(ctx context.Context, tx *sql.Tx, items []WarehouseTransferItem) error
	GetItemsByTransferID(ctx context.Context, transferID uuid.UUID) ([]WarehouseTransferItem, error)
}

type WarehouseUsecase interface {
	Create(ctx context.Context, payload WarehouseCreateRequest) error
	Retrieve(ctx context.Context, id uuid.UUID) (*WareHouseFormatter, error)
	Update(ctx context.Context, id uuid.UUID, payload WarehouseCreateRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetActive(ctx context.Context, id uuid.UUID, isActive bool) error
	GetByShopID(ctx context.Context, shopID uuid.UUID) ([]WareHouseFormatter, error)
}

type WarehouseTransferUsecase interface {
	CreateTransfer(ctx context.Context, req CreateTransferRequest) (*WarehouseTransfer, error)
	UpdateTransferStatus(ctx context.Context, transferID uuid.UUID, req UpdateTransferStatusRequest) error
	GetTransfer(ctx context.Context, transferID uuid.UUID) (*WarehouseTransfer, error)
	GetTransfersByWarehouse(ctx context.Context, warehouseID uuid.UUID, limit, offset int) ([]WarehouseTransfer, int, error)
	ExecuteTransfer(ctx context.Context, transferID uuid.UUID) error
}
