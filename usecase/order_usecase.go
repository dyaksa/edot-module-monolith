package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/dyaksa/warehouse/pkg/paginator"
	"github.com/google/uuid"
)

type orderUsecase struct {
	db                 pqsql.Database
	orderRepo          domain.OrderRepository
	idemRepo           domain.IdempotencyRequestRepository
	orderItemRepo      domain.OrderItemRepository
	reservationRepo    domain.ReservationRepository
	movementRepository domain.MovementRepository
	productStockRepo   domain.ProductStockRepository
	pickWarehouseRepo  domain.WarehouseRepository
}

func (o *orderUsecase) Checkout(ctx context.Context, input domain.CheckoutInput) (*domain.CheckoutOutput, error) {
	if len(input.Items) == 0 {
		return nil, errors.New("order must contain at least one item")
	}

	// Handle TTL configuration - prioritize ReservationMinutes for API convenience
	if input.ReservationMinutes > 0 {
		input.ReservationTTL = time.Duration(input.ReservationMinutes) * time.Minute
	} else if input.ReservationTTL <= 0 {
		input.ReservationTTL = 15 * time.Minute // Default to 15 minutes
	}

	out := &domain.CheckoutOutput{}

	_, err := o.db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		if input.IdemKey != "" {
			isNew, err := o.idemRepo.BeginKey(ctx, tx, input.IdemKey, "checkout", input.PayloadHash)
			if err != nil {
				return nil, err
			}

			if !isNew {
				payloadHash, orderID, responseJSON, exists, err := o.idemRepo.LoadIfExists(ctx, tx, input.IdemKey, "checkout")
				if err != nil {
					return nil, err
				}

				if !exists {
					return nil, domain.ErrIdempotencyConflict
				}

				if payloadHash != input.PayloadHash {
					return nil, domain.ErrIdempotencyConflict
				}

				if orderID != nil && len(responseJSON) > 0 {
					var existingOut domain.CheckoutOutput
					if err = json.Unmarshal(responseJSON, &existingOut); err != nil {
						return nil, err
					}
					*out = existingOut
					return out, nil
				}
			}
		}

		shopId, err := uuid.Parse(input.ShopID)
		if err != nil {
			return nil, errors.New("invalid shop ID format")
		}

		userId, err := uuid.Parse(input.UserID)
		if err != nil {
			return nil, errors.New("invalid user ID format")
		}

		var total int64
		productValidations := make(map[string]bool)

		for _, item := range input.Items {
			if item.Qty <= 0 {
				return nil, errors.New("item quantity must be greater than 0")
			}
			if item.Price <= 0 {
				return nil, errors.New("item price must be greater than 0")
			}

			productId, err := uuid.Parse(item.ProductID)
			if err != nil {
				return nil, errors.New("invalid product ID format: " + item.ProductID)
			}

			if productValidations[item.ProductID] {
				return nil, errors.New("duplicate product in order: " + item.ProductID)
			}
			productValidations[item.ProductID] = true

			_, err = o.pickWarehouseRepo.Pick(ctx, tx, productId, item.Qty, shopId)
			if err != nil {
				if errors.Is(err, domain.ErrOutOfStock) {
					return nil, errors.New("insufficient stock for product: " + item.ProductID)
				}
				return nil, err
			}

			total += int64(item.Qty) * item.Price
		}

		// Step 2: Create order
		order := &domain.Order{
			ID:            uuid.New(),
			ShopID:        shopId,
			UserID:        userId,
			Total:         total,
			Status:        domain.StatusAwaitingPayment,
			ReservedUntil: &time.Time{},
		}

		if err = o.orderRepo.Create(ctx, tx, order); err != nil {
			return nil, err
		}

		var orderItems []domain.OrderItem
		for _, item := range input.Items {
			productId, _ := uuid.Parse(item.ProductID)

			orderItem := domain.OrderItem{
				ID:        uuid.New(),
				OrderID:   order.ID,
				ProductID: productId,
				Qty:       item.Qty,
				Price:     item.Price,
			}
			orderItems = append(orderItems, orderItem)
		}

		if err = o.orderItemRepo.BulkInsert(ctx, tx, orderItems); err != nil {
			return nil, err
		}

		reservationExpiry := time.Now().Add(input.ReservationTTL)
		var reservations []domain.Reservation

		for _, item := range input.Items {
			productId, _ := uuid.Parse(item.ProductID)

			warehouseId, err := o.pickWarehouseRepo.Pick(ctx, tx, productId, item.Qty, shopId)
			if err != nil {
				return nil, err
			}

			stockReserved, err := o.productStockRepo.TryReserveStock(ctx, tx, productId, warehouseId, int32(item.Qty))
			if err != nil {
				return nil, err
			}

			if !stockReserved {
				return nil, domain.ErrOutOfStock
			}

			reservation := domain.Reservation{
				ID:          uuid.New(),
				OrderID:     order.ID,
				ProductID:   productId,
				WarehouseID: warehouseId,
				Qty:         item.Qty,
				Status:      domain.ResvPending,
				ExpiresAt:   reservationExpiry,
			}
			reservations = append(reservations, reservation)

			if err = o.movementRepository.Append(ctx, tx, productId, warehouseId, "RESERVE", item.Qty, "ORDER_CHECKOUT", order.ID); err != nil {
				return nil, err
			}
		}

		if err = o.reservationRepo.CreateMany(ctx, tx, reservations); err != nil {
			return nil, err
		}

		order.ReservedUntil = &reservationExpiry

		out.OrderID = order.ID
		out.Total = order.Total
		out.Status = string(order.Status)
		out.ReservationExpiresAt = reservationExpiry

		if input.IdemKey != "" {
			responseJSON, err := json.Marshal(out)
			if err != nil {
				return nil, err
			}

			if err = o.idemRepo.SaveResponse(ctx, tx, input.IdemKey, "checkout", out.OrderID, responseJSON); err != nil {
				return nil, err
			}
		}

		return out, nil
	})

	return out, err
}

// ConfirmPayment implements domain.OrderUsecase.
func (o *orderUsecase) ConfirmPayment(ctx context.Context, orderID uuid.UUID) error {
	_, err := o.db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		// 1. Get order details
		order, err := o.orderRepo.GetByID(ctx, orderID)
		if err != nil {
			return nil, err
		}

		// 2. Validate order status
		if order.Status != domain.StatusAwaitingPayment {
			return nil, errors.New("order is not in awaiting payment status")
		}

		// 3. Get all pending reservations for this order
		reservations, err := o.reservationRepo.GetByOrderID(ctx, tx, orderID)
		if err != nil {
			return nil, err
		}

		if len(reservations) == 0 {
			return nil, errors.New("no reservations found for order")
		}

		// 4. Check if reservations are still valid (not expired)
		now := time.Now()
		for _, reservation := range reservations {
			if reservation.Status != domain.ResvPending {
				return nil, errors.New("reservation is not in pending status")
			}
			if reservation.ExpiresAt.Before(now) {
				return nil, errors.New("reservation has expired")
			}
		}

		// 5. Commit stock for all reservations
		for _, reservation := range reservations {
			// Commit the stock (reduce on_hand, reduce reserved)
			if err := o.productStockRepo.CommitStock(ctx, tx, reservation.ProductID, reservation.WarehouseID, int32(reservation.Qty)); err != nil {
				return nil, err
			}

			// Log stock movement
			if err := o.movementRepository.Append(ctx, tx, reservation.ProductID, reservation.WarehouseID,
				"COMMIT", reservation.Qty, "ORDER_PAYMENT", orderID); err != nil {
				return nil, err
			}
		}

		// 6. Mark all reservations as committed
		if err := o.reservationRepo.MarkCommitted(ctx, tx, orderID); err != nil {
			return nil, err
		}

		// 7. Update order status to paid
		if err := o.orderRepo.Updatestatus(ctx, orderID, domain.StatusPaid); err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

// CancelOrder implements domain.OrderUsecase.
func (o *orderUsecase) CancelOrder(ctx context.Context, orderID uuid.UUID) error {
	_, err := o.db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		// 1. Get order details
		order, err := o.orderRepo.GetByID(ctx, orderID)
		if err != nil {
			return nil, err
		}

		// 2. Validate order can be cancelled
		if order.Status != domain.StatusAwaitingPayment && order.Status != domain.StatusPending {
			return nil, errors.New("order cannot be cancelled in current status")
		}

		// 3. Get all pending reservations
		reservations, err := o.reservationRepo.GetByOrderID(ctx, tx, orderID)
		if err != nil {
			return nil, err
		}

		// 4. Release stock for all pending reservations
		for _, reservation := range reservations {
			if reservation.Status == domain.ResvPending {
				// Release the reserved stock
				if err := o.productStockRepo.ReleaseStock(ctx, tx, reservation.ProductID, reservation.WarehouseID, int32(reservation.Qty)); err != nil {
					return nil, err
				}

				// Log stock movement
				if err := o.movementRepository.Append(ctx, tx, reservation.ProductID, reservation.WarehouseID,
					"RELEASE", reservation.Qty, "ORDER_CANCELLED", orderID); err != nil {
					return nil, err
				}
			}
		}

		// 5. Mark reservations as released
		if err := o.reservationRepo.MarkReleased(ctx, tx, orderID); err != nil {
			return nil, err
		}

		// 6. Update order status to cancelled
		if err := o.orderRepo.Updatestatus(ctx, orderID, domain.StatusCancelled); err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

// GetOrderDetails implements domain.OrderUsecase.
func (o *orderUsecase) GetOrderDetails(ctx context.Context, orderID uuid.UUID) (*domain.Order, error) {
	return o.orderRepo.GetByID(ctx, orderID)
}

// GetUserOrders implements domain.OrderUsecase.
func (o *orderUsecase) GetUserOrders(ctx context.Context, userID uuid.UUID, pagination paginator.PaginationRequest) (*paginator.PaginationResult[domain.OrderListItem], error) {
	pagination.ValidateAndSetDefault()
	offset := pagination.GetOffset()

	orders, totalCount, err := o.orderRepo.GetByUserID(ctx, userID, pagination.Limit, offset)
	if err != nil {
		return nil, err
	}

	return paginator.NewPaginationResult(orders, totalCount, pagination), nil
}

func NewOrderUsecase(
	db pqsql.Database,
	orderRepo domain.OrderRepository,
	idemRepo domain.IdempotencyRequestRepository,
	orderItemRepo domain.OrderItemRepository,
	reservationRepo domain.ReservationRepository,
	movementRepository domain.MovementRepository,
	productStockRepo domain.ProductStockRepository,
	pickWarehouseRepo domain.WarehouseRepository) domain.OrderUsecase {
	return &orderUsecase{
		db:                 db,
		orderRepo:          orderRepo,
		idemRepo:           idemRepo,
		orderItemRepo:      orderItemRepo,
		reservationRepo:    reservationRepo,
		movementRepository: movementRepository,
		productStockRepo:   productStockRepo,
		pickWarehouseRepo:  pickWarehouseRepo,
	}
}
