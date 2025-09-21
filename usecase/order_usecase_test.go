package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/dyaksa/warehouse/domain"
	mocks "github.com/dyaksa/warehouse/mocks/repository"
	"github.com/dyaksa/warehouse/pkg/paginator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// fakeDB implements pqsql.Database minimal methods used by orderUsecase
type fakeDB struct {
	txErr error
}

func (f *fakeDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}
func (f *fakeDB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeDB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeDB) Transaction(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) (any, error)) (any, error) {
	if f.txErr != nil {
		return nil, f.txErr
	}
	// We don't need a real *sql.Tx; mocks ignore it (they use type assertion). Provide nil.
	return fn(ctx, nil)
}

func TestOrderUsecase_Checkout_Success(t *testing.T) {
	ctx := context.Background()
	db := &fakeDB{}

	orderRepo := mocks.NewMockOrderRepository(t)
	idemRepo := mocks.NewMockIdempotencyRequestRepository(t)
	orderItemRepo := mocks.NewMockOrderItemRepository(t)
	reservationRepo := mocks.NewMockReservationRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	warehouseRepo := mocks.NewMockWarehouseRepository(t)

	uc := NewOrderUsecase(db, orderRepo, idemRepo, orderItemRepo, reservationRepo, movementRepo, productStockRepo, warehouseRepo)

	shopID := uuid.New()
	userID := uuid.New()
	productID := uuid.New()
	warehouseID := uuid.New()

	input := domain.CheckoutInput{
		ShopID:             shopID.String(),
		UserID:             userID.String(),
		Items:              []domain.CheckoutItem{{ProductID: productID.String(), Qty: 2, Price: 500}},
		IdemKey:            "idem-key-1",
		PayloadHash:        "hash123",
		ReservationMinutes: 1,
	}

	// Idempotency begin -> treat as new
	idemRepo.EXPECT().BeginKey(ctx, mock.Anything, input.IdemKey, "checkout", input.PayloadHash).Return(true, nil)

	// Order create
	orderRepo.EXPECT().Create(ctx, mock.Anything, mock.Anything).Return(nil)
	// Order items bulk insert
	orderItemRepo.EXPECT().BulkInsert(ctx, mock.Anything, mock.Anything).Return(nil)

	// Warehouse pick
	warehouseRepo.EXPECT().Pick(ctx, mock.Anything, productID, 2, shopID).Return(warehouseID, nil)
	// Try reserve stock
	productStockRepo.EXPECT().TryReserveStock(ctx, mock.Anything, productID, warehouseID, int32(2)).Return(true, nil)
	// Movement append
	movementRepo.EXPECT().Append(ctx, mock.Anything, productID, warehouseID, "RESERVE", 2, "ORDER_CHECKOUT", mock.Anything).Return(nil)
	// Reservation create
	reservationRepo.EXPECT().CreateMany(ctx, mock.Anything, mock.Anything).Return(nil)
	// Save response
	idemRepo.EXPECT().SaveResponse(ctx, mock.Anything, input.IdemKey, "checkout", mock.Anything, mock.Anything).Return(nil)

	out, err := uc.Checkout(ctx, input)
	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, int64(1000), out.Total)
	assert.Equal(t, string(domain.StatusAwaitingPayment), out.Status)
	assert.WithinDuration(t, time.Now().Add(1*time.Minute), out.ReservationExpiresAt, 5*time.Second)
}

func TestOrderUsecase_Checkout_EmptyItems(t *testing.T) {
	ctx := context.Background()
	db := &fakeDB{}
	uc := NewOrderUsecase(db, nil, nil, nil, nil, nil, nil, nil)

	out, err := uc.Checkout(ctx, domain.CheckoutInput{ShopID: uuid.New().String(), UserID: uuid.New().String(), Items: []domain.CheckoutItem{}})
	assert.Error(t, err)
	assert.Nil(t, out)
}

func TestOrderUsecase_GetUserOrders(t *testing.T) {
	ctx := context.Background()
	db := &fakeDB{}
	orderRepo := mocks.NewMockOrderRepository(t)
	uc := NewOrderUsecase(db, orderRepo, nil, nil, nil, nil, nil, nil)
	userID := uuid.New()
	orders := []domain.OrderListItem{{ID: uuid.New(), Total: 1000, Status: string(domain.StatusAwaitingPayment)}}
	orderRepo.EXPECT().GetByUserID(ctx, userID, 10, 0).Return(orders, 1, nil)
	p := paginator.PaginationRequest{Page: 1, Limit: 10}
	res, err := uc.GetUserOrders(ctx, userID, p)
	assert.NoError(t, err)
	assert.Equal(t, 1, res.TotalItems)
}
