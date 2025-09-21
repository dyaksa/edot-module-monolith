package domain

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/dyaksa/warehouse/pkg/paginator"
	"github.com/google/uuid"
)

type OrderStatus string

const (
	StatusPending         OrderStatus = "PENDING"
	StatusAwaitingPayment OrderStatus = "AWAITING_PAYMENT"
	StatusPaid            OrderStatus = "PAID"
	StatusCancelled       OrderStatus = "CANCELLED"
	StatusExpired         OrderStatus = "EXPIRED"
	StatusFulfilled       OrderStatus = "FULFILLED"
)

var (
	ErrInvalidStatusTransition = errors.New("invalid status transition")
	ErrOutOfStock              = errors.New("out of stock")
	ErrIdempotencyConflict     = errors.New("idempotency conflict")
)

type Order struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	ShopID        uuid.UUID
	Status        OrderStatus
	Items         []OrderItem
	Total         int64
	ReservedUntil *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type OrderItem struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	ProductID uuid.UUID
	Qty       int
	Price     int64
}

type ReservationStatus string

const (
	ResvPending   ReservationStatus = "PENDING"
	ResvCommitted ReservationStatus = "COMMITTED"
	ResvReleased  ReservationStatus = "RELEASED"
	ResvExpired   ReservationStatus = "EXPIRED"
)

type Reservation struct {
	ID          uuid.UUID
	OrderID     uuid.UUID
	ProductID   uuid.UUID
	WarehouseID uuid.UUID
	Qty         int
	Status      ReservationStatus
	ExpiresAt   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type CheckoutItem struct {
	ProductID string `json:"product_id" binding:"required"`
	Qty       int    `json:"qty" binding:"required,gt=0"`
	Price     int64  `json:"price" binding:"required,gt=0"`
}

type CheckoutInput struct {
	ShopID             string         `json:"shop_id" binding:"required"`
	Items              []CheckoutItem `json:"items" binding:"required,dive,required"`
	UserID             string         `json:"user_id" binding:"required"`
	IdemKey            string         `json:"idem_key"`
	PayloadHash        string         `json:"payload_hash"`
	ReservationTTL     time.Duration  `json:"reservation_ttl"`     // in seconds (for internal use)
	ReservationMinutes int            `json:"reservation_minutes"` // in minutes (for API convenience)
}

type CheckoutOutput struct {
	OrderID              uuid.UUID `json:"order_id"`
	ReservationExpiresAt time.Time `json:"reservation_expires_at"`
	Total                int64     `json:"total"`
	Status               string    `json:"status"`
}

type OrderListItem struct {
	ID                   uuid.UUID  `json:"order_id"`
	Total                int64      `json:"total"`
	Status               string     `json:"status"`
	ItemCount            int        `json:"item_count"`
	ReservationExpiresAt *time.Time `json:"reservation_expires_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

type OrderRepository interface {
	Create(ctx context.Context, tx *sql.Tx, o *Order) error
	Updatestatus(ctx context.Context, orderID uuid.UUID, status OrderStatus) error
	GetByID(ctx context.Context, orderID uuid.UUID) (*Order, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]OrderListItem, int, error)
}

type OrderItemRepository interface {
	BulkInsert(ctx context.Context, tx *sql.Tx, items []OrderItem) error
}

type ReservationRepository interface {
	CreateMany(ctx context.Context, tx *sql.Tx, reservations []Reservation) error
	PickExpiredForUpdate(ctx context.Context, tx *sql.Tx, limit int) ([]Reservation, error)
	MarkExpired(ctx context.Context, tx *sql.Tx, id uuid.UUID) error
	MarkCommitted(ctx context.Context, tx *sql.Tx, orderID uuid.UUID) error
	MarkReleased(ctx context.Context, tx *sql.Tx, orderID uuid.UUID) error
	PendingCountByOrder(ctx context.Context, tx *sql.Tx, orderID uuid.UUID) (int, error)
	Retrieve(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*Reservation, error)
	GetByOrderID(ctx context.Context, tx *sql.Tx, orderID uuid.UUID) ([]Reservation, error)
}

type MovementRepository interface {
	Append(ctx context.Context, tx *sql.Tx, productID, warehouseID uuid.UUID, typ string, qty int, refType string, refID uuid.UUID) error
}

type OrderUsecase interface {
	Checkout(ctx context.Context, input CheckoutInput) (*CheckoutOutput, error)
	ConfirmPayment(ctx context.Context, orderID uuid.UUID) error
	CancelOrder(ctx context.Context, orderID uuid.UUID) error
	GetOrderDetails(ctx context.Context, orderID uuid.UUID) (*Order, error)
	GetUserOrders(ctx context.Context, userID uuid.UUID, pagination paginator.PaginationRequest) (*paginator.PaginationResult[OrderListItem], error)
}
