package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/dyaksa/warehouse/domain"
	mocks "github.com/dyaksa/warehouse/mocks/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// reuse fakeDB from order tests if needed; redefine simplified variant here
type fakeDBStock struct{}

func (f *fakeDBStock) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}
func (f *fakeDBStock) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeDBStock) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeDBStock) Transaction(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) (any, error)) (any, error) {
	return fn(ctx, nil)
}

func TestStockRelease_ProcessExpiredReservations_Success(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBStock{}
	reservationRepo := mocks.NewMockReservationRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	orderRepo := mocks.NewMockOrderRepository(t)

	uc := NewStockReleaseUsecase(db, reservationRepo, productStockRepo, movementRepo, orderRepo)

	orderID := uuid.New()
	productID := uuid.New()
	warehouseID := uuid.New()
	res1 := domain.Reservation{ID: uuid.New(), OrderID: orderID, ProductID: productID, WarehouseID: warehouseID, Qty: 3}
	expired := []domain.Reservation{res1}

	reservationRepo.EXPECT().PickExpiredForUpdate(ctx, mock.Anything, 50).Return(expired, nil)
	productStockRepo.EXPECT().ReleaseStock(ctx, mock.Anything, productID, warehouseID, int32(3)).Return(nil)
	movementRepo.EXPECT().Append(ctx, mock.Anything, productID, warehouseID, "RELEASE", 3, "RESERVATION_EXPIRED", res1.ID).Return(nil)
	reservationRepo.EXPECT().MarkExpired(ctx, mock.Anything, res1.ID).Return(nil)
	reservationRepo.EXPECT().PendingCountByOrder(ctx, mock.Anything, orderID).Return(0, nil)
	orderRepo.EXPECT().Updatestatus(ctx, orderID, domain.StatusExpired).Return(nil)

	err := uc.ProcessExpiredReservations(ctx, 50)
	assert.NoError(t, err)
}

func TestStockRelease_ProcessExpiredReservations_PendingRemain(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBStock{}
	reservationRepo := mocks.NewMockReservationRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	orderRepo := mocks.NewMockOrderRepository(t)

	uc := NewStockReleaseUsecase(db, reservationRepo, productStockRepo, movementRepo, orderRepo)

	orderID := uuid.New()
	productID := uuid.New()
	warehouseID := uuid.New()
	res1 := domain.Reservation{ID: uuid.New(), OrderID: orderID, ProductID: productID, WarehouseID: warehouseID, Qty: 2}
	expired := []domain.Reservation{res1}

	reservationRepo.EXPECT().PickExpiredForUpdate(ctx, mock.Anything, 10).Return(expired, nil)
	productStockRepo.EXPECT().ReleaseStock(ctx, mock.Anything, productID, warehouseID, int32(2)).Return(nil)
	movementRepo.EXPECT().Append(ctx, mock.Anything, productID, warehouseID, "RELEASE", 2, "RESERVATION_EXPIRED", res1.ID).Return(nil)
	reservationRepo.EXPECT().MarkExpired(ctx, mock.Anything, res1.ID).Return(nil)
	reservationRepo.EXPECT().PendingCountByOrder(ctx, mock.Anything, orderID).Return(1, nil) // still pending others
	// orderRepo.Updatestatus should NOT be called; absence is asserted by mock expectations auto-verify

	err := uc.ProcessExpiredReservations(ctx, 10)
	assert.NoError(t, err)
}

func TestStockRelease_ReleaseReservationStock_Success(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBStock{}
	reservationRepo := mocks.NewMockReservationRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	orderRepo := mocks.NewMockOrderRepository(t)

	uc := NewStockReleaseUsecase(db, reservationRepo, productStockRepo, movementRepo, orderRepo)

	reservation := domain.Reservation{ID: uuid.New(), ProductID: uuid.New(), WarehouseID: uuid.New(), Qty: 5}
	productStockRepo.EXPECT().ReleaseStock(ctx, mock.Anything, reservation.ProductID, reservation.WarehouseID, int32(reservation.Qty)).Return(nil)
	movementRepo.EXPECT().Append(ctx, mock.Anything, reservation.ProductID, reservation.WarehouseID, "RELEASE", reservation.Qty, "RESERVATION_EXPIRED", reservation.ID).Return(nil)

	err := uc.ReleaseReservationStock(ctx, reservation)
	assert.NoError(t, err)
}

func TestStockRelease_ProcessExpiredReservations_ErrorOnPick(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBStock{}
	reservationRepo := mocks.NewMockReservationRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	orderRepo := mocks.NewMockOrderRepository(t)

	uc := NewStockReleaseUsecase(db, reservationRepo, productStockRepo, movementRepo, orderRepo)

	expected := errors.New("pick failed")
	reservationRepo.EXPECT().PickExpiredForUpdate(ctx, mock.Anything, 5).Return(nil, expected)

	err := uc.ProcessExpiredReservations(ctx, 5)
	assert.ErrorIs(t, err, expected)
}

func TestStockRelease_ProcessExpiredReservations_ErrorOnReleaseStock(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBStock{}
	reservationRepo := mocks.NewMockReservationRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	orderRepo := mocks.NewMockOrderRepository(t)

	uc := NewStockReleaseUsecase(db, reservationRepo, productStockRepo, movementRepo, orderRepo)

	orderID := uuid.New()
	productID := uuid.New()
	warehouseID := uuid.New()
	res1 := domain.Reservation{ID: uuid.New(), OrderID: orderID, ProductID: productID, WarehouseID: warehouseID, Qty: 1}
	expired := []domain.Reservation{res1}

	reservationRepo.EXPECT().PickExpiredForUpdate(ctx, mock.Anything, 3).Return(expired, nil)
	productStockRepo.EXPECT().ReleaseStock(ctx, mock.Anything, productID, warehouseID, int32(1)).Return(errors.New("release fail"))

	err := uc.ProcessExpiredReservations(ctx, 3)
	assert.Error(t, err)
}

// Additional edge test: no expired reservations -> should just pass
func TestStockRelease_ProcessExpiredReservations_NoExpired(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBStock{}
	reservationRepo := mocks.NewMockReservationRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	orderRepo := mocks.NewMockOrderRepository(t)

	uc := NewStockReleaseUsecase(db, reservationRepo, productStockRepo, movementRepo, orderRepo)

	reservationRepo.EXPECT().PickExpiredForUpdate(ctx, mock.Anything, 20).Return([]domain.Reservation{}, nil)
	// No further calls expected
	err := uc.ProcessExpiredReservations(ctx, 20)
	assert.NoError(t, err)
	// time-based no action side effects
	_ = time.Now() // keep time import used when no other test uses time
}
