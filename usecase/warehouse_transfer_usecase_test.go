package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/dyaksa/warehouse/domain"
	mocks "github.com/dyaksa/warehouse/mocks/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type fakeDBTransfer struct{}

func (f *fakeDBTransfer) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return nil
}
func (f *fakeDBTransfer) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeDBTransfer) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, errors.New("not implemented")
}
func (f *fakeDBTransfer) Transaction(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) (any, error)) (any, error) {
	return fn(ctx, nil)
}

func TestWarehouseTransfer_CreateTransfer_Success(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBTransfer{}
	transferRepo := mocks.NewMockWarehouseTransferRepository(t)
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	uc := NewWarehouseTransferUsecase(db, transferRepo, warehouseRepo, productStockRepo, movementRepo)

	shopID := uuid.New()
	fromW := &domain.WareHouse{ID: uuid.New(), ShopID: shopID, IsActive: true}
	toW := &domain.WareHouse{ID: uuid.New(), ShopID: shopID, IsActive: true}
	productID := uuid.New()

	warehouseRepo.EXPECT().Retrieve(ctx, fromW.ID).Return(fromW, nil)
	// to warehouse still retrieved per code path even if from is inactive
	warehouseRepo.EXPECT().Retrieve(ctx, toW.ID).Return(toW, nil)
	transferRepo.EXPECT().Create(ctx, mock.Anything, mock.Anything).Return(nil)
	transferRepo.EXPECT().CreateItems(ctx, mock.Anything, mock.Anything).Return(nil)

	req := domain.CreateTransferRequest{
		FromWarehouseID: fromW.ID.String(),
		ToWarehouseID:   toW.ID.String(),
		Items:           []domain.CreateTransferItemRequest{{ProductID: productID.String(), Qty: 5}},
	}

	transfer, err := uc.CreateTransfer(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, transfer)
	assert.Equal(t, domain.TransferStatusRequested, transfer.Status)
	assert.Len(t, transfer.Items, 1)
}

func TestWarehouseTransfer_CreateTransfer_InvalidWarehouseID(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBTransfer{}
	uc := NewWarehouseTransferUsecase(db, nil, nil, nil, nil)

	_, err := uc.CreateTransfer(ctx, domain.CreateTransferRequest{FromWarehouseID: "bad", ToWarehouseID: uuid.New().String(), Items: []domain.CreateTransferItemRequest{}})
	assert.Error(t, err)
}

func TestWarehouseTransfer_CreateTransfer_SameWarehouse(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBTransfer{}
	id := uuid.New()
	uc := NewWarehouseTransferUsecase(db, nil, nil, nil, nil)
	_, err := uc.CreateTransfer(ctx, domain.CreateTransferRequest{FromWarehouseID: id.String(), ToWarehouseID: id.String(), Items: []domain.CreateTransferItemRequest{}})
	assert.Error(t, err)
}

func TestWarehouseTransfer_CreateTransfer_InactiveWarehouse(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBTransfer{}
	transferRepo := mocks.NewMockWarehouseTransferRepository(t)
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	uc := NewWarehouseTransferUsecase(db, transferRepo, warehouseRepo, productStockRepo, movementRepo)

	shopID := uuid.New()
	fromW := &domain.WareHouse{ID: uuid.New(), ShopID: shopID, IsActive: false}
	toW := &domain.WareHouse{ID: uuid.New(), ShopID: shopID, IsActive: true}

	warehouseRepo.EXPECT().Retrieve(ctx, fromW.ID).Return(fromW, nil)

	req := domain.CreateTransferRequest{FromWarehouseID: fromW.ID.String(), ToWarehouseID: toW.ID.String(), Items: []domain.CreateTransferItemRequest{}}
	_, err := uc.CreateTransfer(ctx, req)
	assert.Error(t, err)
}

func TestWarehouseTransfer_CreateTransfer_DifferentShop(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBTransfer{}
	transferRepo := mocks.NewMockWarehouseTransferRepository(t)
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	uc := NewWarehouseTransferUsecase(db, transferRepo, warehouseRepo, productStockRepo, movementRepo)

	fromW := &domain.WareHouse{ID: uuid.New(), ShopID: uuid.New(), IsActive: true}
	toW := &domain.WareHouse{ID: uuid.New(), ShopID: uuid.New(), IsActive: true}

	warehouseRepo.EXPECT().Retrieve(ctx, fromW.ID).Return(fromW, nil)
	warehouseRepo.EXPECT().Retrieve(ctx, toW.ID).Return(toW, nil)

	req := domain.CreateTransferRequest{FromWarehouseID: fromW.ID.String(), ToWarehouseID: toW.ID.String(), Items: []domain.CreateTransferItemRequest{}}
	_, err := uc.CreateTransfer(ctx, req)
	assert.Error(t, err)
}

func TestWarehouseTransfer_CreateTransfer_InactiveDestination(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBTransfer{}
	transferRepo := mocks.NewMockWarehouseTransferRepository(t)
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	uc := NewWarehouseTransferUsecase(db, transferRepo, warehouseRepo, productStockRepo, movementRepo)

	shopID := uuid.New()
	fromW := &domain.WareHouse{ID: uuid.New(), ShopID: shopID, IsActive: true}
	toW := &domain.WareHouse{ID: uuid.New(), ShopID: shopID, IsActive: false}

	warehouseRepo.EXPECT().Retrieve(ctx, fromW.ID).Return(fromW, nil)
	warehouseRepo.EXPECT().Retrieve(ctx, toW.ID).Return(toW, nil)

	req := domain.CreateTransferRequest{FromWarehouseID: fromW.ID.String(), ToWarehouseID: toW.ID.String(), Items: []domain.CreateTransferItemRequest{}}
	_, err := uc.CreateTransfer(ctx, req)
	assert.Error(t, err)
}

func TestWarehouseTransfer_ExecuteTransfer_Success(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBTransfer{}
	transferRepo := mocks.NewMockWarehouseTransferRepository(t)
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	uc := NewWarehouseTransferUsecase(db, transferRepo, warehouseRepo, productStockRepo, movementRepo)

	transferID := uuid.New()
	fromW := uuid.New()
	toW := uuid.New()
	productID := uuid.New()
	items := []domain.WarehouseTransferItem{{ID: uuid.New(), TransferID: transferID, ProductID: productID, Qty: 3}}
	transfer := &domain.WarehouseTransfer{ID: transferID, FromWarehouseID: fromW, ToWarehouseID: toW, Status: domain.TransferStatusRequested, Items: items}

	transferRepo.EXPECT().GetByID(ctx, transferID).Return(transfer, nil)
	productStockRepo.EXPECT().TryReserveStock(ctx, mock.Anything, productID, fromW, int32(3)).Return(true, nil)
	movementRepo.EXPECT().Append(ctx, mock.Anything, productID, fromW, mock.Anything, 3, mock.Anything, transferID).Return(nil)
	productStockRepo.EXPECT().CommitStock(ctx, mock.Anything, productID, fromW, int32(3)).Return(nil)
	transferRepo.EXPECT().UpdateStatus(ctx, mock.Anything, transferID, domain.TransferStatusInTransit).Return(nil)
	movementRepo.EXPECT().Append(ctx, mock.Anything, productID, toW, mock.Anything, 3, mock.Anything, transferID).Return(nil)
	productStockRepo.EXPECT().AddStock(ctx, mock.Anything, productID, toW, int32(3)).Return(nil)
	transferRepo.EXPECT().UpdateStatus(ctx, mock.Anything, transferID, domain.TransferStatusCompleted).Return(nil)

	err := uc.ExecuteTransfer(ctx, transferID)
	assert.NoError(t, err)
}

func TestWarehouseTransfer_ExecuteTransfer_InsufficientStock(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBTransfer{}
	transferRepo := mocks.NewMockWarehouseTransferRepository(t)
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	uc := NewWarehouseTransferUsecase(db, transferRepo, warehouseRepo, productStockRepo, movementRepo)

	transferID := uuid.New()
	fromW := uuid.New()
	toW := uuid.New()
	productID := uuid.New()
	items := []domain.WarehouseTransferItem{{ID: uuid.New(), TransferID: transferID, ProductID: productID, Qty: 4}}
	transfer := &domain.WarehouseTransfer{ID: transferID, FromWarehouseID: fromW, ToWarehouseID: toW, Status: domain.TransferStatusRequested, Items: items}

	transferRepo.EXPECT().GetByID(ctx, transferID).Return(transfer, nil)
	productStockRepo.EXPECT().TryReserveStock(ctx, mock.Anything, productID, fromW, int32(4)).Return(false, nil)

	err := uc.ExecuteTransfer(ctx, transferID)
	assert.Error(t, err)
}

func TestWarehouseTransfer_UpdateTransferStatus_ExecuteInTransit(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBTransfer{}
	transferRepo := mocks.NewMockWarehouseTransferRepository(t)
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	uc := NewWarehouseTransferUsecase(db, transferRepo, warehouseRepo, productStockRepo, movementRepo)

	transferID := uuid.New()
	fromW := uuid.New()
	toW := uuid.New()
	productID := uuid.New()
	items := []domain.WarehouseTransferItem{{ID: uuid.New(), TransferID: transferID, ProductID: productID, Qty: 1}}
	transfer := &domain.WarehouseTransfer{ID: transferID, FromWarehouseID: fromW, ToWarehouseID: toW, Status: domain.TransferStatusApproved, Items: items}

	// UpdateTransferStatus will call GetByID then ExecuteTransfer path
	transferRepo.EXPECT().GetByID(ctx, transferID).Return(transfer, nil).Once()
	// ExecuteTransfer internals
	transferRepo.EXPECT().GetByID(ctx, transferID).Return(transfer, nil).Once()
	productStockRepo.EXPECT().TryReserveStock(ctx, mock.Anything, productID, fromW, int32(1)).Return(true, nil)
	movementRepo.EXPECT().Append(ctx, mock.Anything, productID, fromW, mock.Anything, 1, mock.Anything, transferID).Return(nil)
	productStockRepo.EXPECT().CommitStock(ctx, mock.Anything, productID, fromW, int32(1)).Return(nil)
	transferRepo.EXPECT().UpdateStatus(ctx, mock.Anything, transferID, domain.TransferStatusInTransit).Return(nil)
	movementRepo.EXPECT().Append(ctx, mock.Anything, productID, toW, mock.Anything, 1, mock.Anything, transferID).Return(nil)
	productStockRepo.EXPECT().AddStock(ctx, mock.Anything, productID, toW, int32(1)).Return(nil)
	transferRepo.EXPECT().UpdateStatus(ctx, mock.Anything, transferID, domain.TransferStatusCompleted).Return(nil)

	err := uc.UpdateTransferStatus(ctx, transferID, domain.UpdateTransferStatusRequest{Status: domain.TransferStatusInTransit})
	assert.NoError(t, err)
}

func TestWarehouseTransfer_UpdateTransferStatus_InvalidTransition(t *testing.T) {
	ctx := context.Background()
	db := &fakeDBTransfer{}
	transferRepo := mocks.NewMockWarehouseTransferRepository(t)
	warehouseRepo := mocks.NewMockWarehouseRepository(t)
	productStockRepo := mocks.NewMockProductStockRepository(t)
	movementRepo := mocks.NewMockMovementRepository(t)
	uc := NewWarehouseTransferUsecase(db, transferRepo, warehouseRepo, productStockRepo, movementRepo)

	transferID := uuid.New()
	transfer := &domain.WarehouseTransfer{ID: transferID, Status: domain.TransferStatusCompleted}
	transferRepo.EXPECT().GetByID(ctx, transferID).Return(transfer, nil)

	err := uc.UpdateTransferStatus(ctx, transferID, domain.UpdateTransferStatusRequest{Status: domain.TransferStatusApproved})
	assert.Error(t, err)
}
