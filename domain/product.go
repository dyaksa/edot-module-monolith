package domain

import (
	"context"
	"database/sql"
	"time"

	"github.com/dyaksa/warehouse/pkg/paginator"
	"github.com/google/uuid"
)

type Product struct {
	ID   uuid.UUID
	SKU  string
	Name string
}

type ProductStock struct {
	ID          uuid.UUID
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	OnHand      int32
	Reserved    int32
	UpdatedAt   time.Time
}

// CreateProductRequest represents the request payload for creating a new product
type CreateProductRequest struct {
	WarehouseID string `json:"warehouse_id" binding:"required" example:"550e8400-e29b-41d4-a716-446655440000" description:"UUID of the warehouse where the product will be stored"`
	SKU         string `json:"sku" binding:"required" example:"PROD-001" description:"Stock Keeping Unit - unique product identifier"`
	Name        string `json:"name" binding:"required" example:"Sample Product" description:"Product name"`
	OnHand      int32  `json:"on_hand" example:"100" description:"Initial stock quantity available"`
}

// RetrieveProduct represents the response payload for product retrieval
type RetrieveProduct struct {
	ID            uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" description:"Product UUID"`
	SKU           string    `json:"sku" example:"PROD-001" description:"Stock Keeping Unit"`
	Name          string    `json:"name" example:"Sample Product" description:"Product name"`
	WarehouseName string    `json:"warehouse_name,omitempty" example:"Main Warehouse" description:"Name of the warehouse where product is stored"`
	Available     int32     `json:"available" example:"85" description:"Available stock quantity (on_hand - reserved)"`
}

type ProductRepository interface {
	Create(ctx context.Context, product *Product) (uuid.UUID, error)
	RetrieveAll(ctx context.Context, limit, offset int) ([]RetrieveProduct, error)
}

type ProductStockRepository interface {
	Create(ctx context.Context, productStock *ProductStock) (uuid.UUID, error)
	TryReserveStock(ctx context.Context, tx *sql.Tx, productID, warehouseID uuid.UUID, quantity int32) (bool, error)
	ReleaseStock(ctx context.Context, tx *sql.Tx, productID, warehouseID uuid.UUID, quantity int32) error
	CommitStock(ctx context.Context, tx *sql.Tx, productID, warehouseID uuid.UUID, quantity int32) error
	AddStock(ctx context.Context, tx *sql.Tx, productID, warehouseID uuid.UUID, quantity int32) error
}

type ProductUsecase interface {
	Create(ctx context.Context, payload CreateProductRequest) error
	RetrieveAll(ctx context.Context, pagination paginator.PaginationRequest) (*paginator.PaginationResult[RetrieveProduct], error)
}
