package usecase

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/dyaksa/warehouse/pkg/errx"
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
		return nil, errx.E(errx.CodeValidation, "order must contain at least one item", errx.Op("OrderUsecase.Checkout"))
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
				return nil, errx.E(errx.CodeInternal, "failed to begin idempotency key", errx.Op("OrderUsecase.Checkout"), err)
			}

			if !isNew {
				payloadHash, orderID, responseJSON, exists, err := o.idemRepo.LoadIfExists(ctx, tx, input.IdemKey, "checkout")
				if err != nil {
					return nil, errx.E(errx.CodeInternal, "failed to load idempotency key", errx.Op("OrderUsecase.Checkout"), err)
				}

				if !exists {
					return nil, errx.E(errx.CodeInternal, "idempotency key not found", errx.Op("OrderUsecase.Checkout"), domain.ErrIdempotencyConflict)
				}

				if payloadHash != input.PayloadHash {
					return nil, errx.E(errx.CodeInternal, "idempotency key payload hash mismatch", errx.Op("OrderUsecase.Checkout"), domain.ErrIdempotencyConflict)
				}

				if orderID != nil && len(responseJSON) > 0 {
					var existingOut domain.CheckoutOutput
					if err = json.Unmarshal(responseJSON, &existingOut); err != nil {
						return nil, errx.E(errx.CodeInternal, "failed to unmarshal idempotent response", errx.Op("OrderUsecase.Checkout"), err)
					}
					*out = existingOut
					return out, nil
				}
			}
		}

		shopId, err := uuid.Parse(input.ShopID)
		if err != nil {
			return nil, errx.E(errx.CodeValidation, "invalid shop ID format", errx.Op("OrderUsecase.Checkout"), err)
		}

		userId, err := uuid.Parse(input.UserID)
		if err != nil {
			return nil, errx.E(errx.CodeValidation, "invalid user ID format", errx.Op("OrderUsecase.Checkout"), err)
		}

		var total int64
		productValidations := make(map[string]bool)

		for _, item := range input.Items {
			if item.Qty <= 0 {
				return nil, errx.E(errx.CodeValidation, "item quantity must be greater than 0", errx.Op("OrderUsecase.Checkout"))
			}
			if item.Price <= 0 {
				return nil, errx.E(errx.CodeValidation, "item price must be greater than 0", errx.Op("OrderUsecase.Checkout"))
			}

			productId, err := uuid.Parse(item.ProductID)
			if err != nil {
				return nil, errx.E(errx.CodeValidation, "invalid product ID format", errx.Op("OrderUsecase.Checkout"), err)
			}

			if productValidations[item.ProductID] {
				return nil, errx.E(errx.CodeValidation, "duplicate product in order", errx.Op("OrderUsecase.Checkout"), errors.New(item.ProductID))
			}
			productValidations[item.ProductID] = true

			_, err = o.pickWarehouseRepo.Pick(ctx, tx, productId, item.Qty, shopId)
			if err != nil {
				if errors.Is(err, domain.ErrOutOfStock) {
					return nil, errx.E(errx.CodeValidation, "insufficient stock for product", errx.Op("OrderUsecase.Checkout"), errors.New(item.ProductID))
				}
				return nil, errx.E(errx.CodeInternal, "failed to validate product stock", errx.Op("OrderUsecase.Checkout"), err)
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
			return nil, errx.E(errx.CodeInternal, "failed to create order", errx.Op("OrderUsecase.Checkout"), err)
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
				return nil, errx.E(errx.CodeInternal, "failed to pick warehouse for product", errx.Op("OrderUsecase.Checkout"), err)
			}

			stockReserved, err := o.productStockRepo.TryReserveStock(ctx, tx, productId, warehouseId, int32(item.Qty))
			if err != nil {
				return nil, errx.E(errx.CodeInternal, "failed to reserve stock for product", errx.Op("OrderUsecase.Checkout"), err)
			}

			if !stockReserved {
				return nil, errx.E(errx.CodeValidation, "insufficient stock to reserve for product", errx.Op("OrderUsecase.Checkout"), domain.ErrOutOfStock)
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
				return nil, errx.E(errx.CodeInternal, "failed to log stock reservation", errx.Op("OrderUsecase.Checkout"), err)
			}
		}

		if err = o.reservationRepo.CreateMany(ctx, tx, reservations); err != nil {
			return nil, errx.E(errx.CodeInternal, "failed to create reservations", errx.Op("OrderUsecase.Checkout"), err)
		}

		order.ReservedUntil = &reservationExpiry

		out.OrderID = order.ID
		out.Total = order.Total
		out.Status = string(order.Status)
		out.ReservationExpiresAt = reservationExpiry

		if input.IdemKey != "" {
			responseJSON, err := json.Marshal(out)
			if err != nil {
				return nil, errx.E(errx.CodeInternal, "failed to marshal response", errx.Op("OrderUsecase.Checkout"), err)
			}

			if err = o.idemRepo.SaveResponse(ctx, tx, input.IdemKey, "checkout", out.OrderID, responseJSON); err != nil {
				return nil, errx.E(errx.CodeInternal, "failed to save response", errx.Op("OrderUsecase.Checkout"), err)
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
			return nil, errx.E(errx.CodeInternal, "failed to get order", errx.Op("OrderUsecase.ConfirmPayment"), err)
		}

		// 2. Validate order status
		if order.Status != domain.StatusAwaitingPayment {
			return nil, errx.E(errx.CodeValidation, "order cannot be confirmed in current status", errx.Op("OrderUsecase.ConfirmPayment"))
		}

		// 3. Get all pending reservations for this order
		reservations, err := o.reservationRepo.GetByOrderID(ctx, tx, orderID)
		if err != nil {
			return nil, errx.E(errx.CodeInternal, "failed to get reservations", errx.Op("OrderUsecase.ConfirmPayment"), err)
		}

		if len(reservations) == 0 {
			return nil, errx.E(errx.CodeValidation, "no reservations found for order", errx.Op("OrderUsecase.ConfirmPayment"))
		}

		// 4. Check if reservations are still valid (not expired)
		now := time.Now()
		for _, reservation := range reservations {
			if reservation.Status != domain.ResvPending {
				return nil, errx.E(errx.CodeValidation, "reservation is not in pending status", errx.Op("OrderUsecase.ConfirmPayment"))
			}
			if reservation.ExpiresAt.Before(now) {
				return nil, errx.E(errx.CodeValidation, "reservation has expired", errx.Op("OrderUsecase.ConfirmPayment"))
			}
		}

		// 5. Commit stock for all reservations
		for _, reservation := range reservations {
			// Commit the stock (reduce on_hand, reduce reserved)
			if err := o.productStockRepo.CommitStock(ctx, tx, reservation.ProductID, reservation.WarehouseID, int32(reservation.Qty)); err != nil {
				return nil, errx.E(errx.CodeInternal, "failed to commit stock for reservation", errx.Op("OrderUsecase.ConfirmPayment"), err)
			}

			// Log stock movement
			if err := o.movementRepository.Append(ctx, tx, reservation.ProductID, reservation.WarehouseID,
				"COMMIT", reservation.Qty, "ORDER_PAYMENT", orderID); err != nil {
				return nil, errx.E(errx.CodeInternal, "failed to log stock commit", errx.Op("OrderUsecase.ConfirmPayment"), err)
			}
		}

		// 6. Mark all reservations as committed
		if err := o.reservationRepo.MarkCommitted(ctx, tx, orderID); err != nil {
			return nil, errx.E(errx.CodeInternal, "failed to mark reservations as committed", errx.Op("OrderUsecase.ConfirmPayment"), err)
		}

		// 7. Update order status to paid
		if err := o.orderRepo.Updatestatus(ctx, orderID, domain.StatusPaid); err != nil {
			return nil, errx.E(errx.CodeInternal, "failed to update order status", errx.Op("OrderUsecase.ConfirmPayment"), err)
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
