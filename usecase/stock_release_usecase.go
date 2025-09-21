package usecase

import (
	"context"
	"database/sql"
	"log"

	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
)

type StockReleaseUsecase interface {
	ProcessExpiredReservations(ctx context.Context, batchSize int) error
	ReleaseReservationStock(ctx context.Context, reservation domain.Reservation) error
}

type stockReleaseUsecase struct {
	db               pqsql.Database
	reservationRepo  domain.ReservationRepository
	productStockRepo domain.ProductStockRepository
	movementRepo     domain.MovementRepository
	orderRepo        domain.OrderRepository
}

// ProcessExpiredReservations implements StockReleaseUsecase.
func (s *stockReleaseUsecase) ProcessExpiredReservations(ctx context.Context, batchSize int) error {
	_, err := s.db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		// Pick expired reservations for update
		expiredReservations, err := s.reservationRepo.PickExpiredForUpdate(ctx, tx, batchSize)
		if err != nil {
			return nil, err
		}

		log.Printf("Found %d expired reservations to process", len(expiredReservations))

		for _, reservation := range expiredReservations {
			// Release stock for each expired reservation
			if err := s.releaseStockForReservation(ctx, tx, reservation); err != nil {
				log.Printf("Error releasing stock for reservation %s: %v", reservation.ID, err)
				return nil, err
			}

			// Mark reservation as expired
			if err := s.reservationRepo.MarkExpired(ctx, tx, reservation.ID); err != nil {
				log.Printf("Error marking reservation %s as expired: %v", reservation.ID, err)
				return nil, err
			}

			log.Printf("Successfully released stock and marked reservation %s as expired", reservation.ID)
		}

		// Check if all reservations for orders are expired and update order status
		if err := s.updateOrderStatusIfAllReservationsExpired(ctx, tx, expiredReservations); err != nil {
			return nil, err
		}

		return nil, nil
	})

	return err
}

// ReleaseReservationStock implements StockReleaseUsecase.
func (s *stockReleaseUsecase) ReleaseReservationStock(ctx context.Context, reservation domain.Reservation) error {
	_, err := s.db.Transaction(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		return nil, s.releaseStockForReservation(ctx, tx, reservation)
	})
	return err
}

func (s *stockReleaseUsecase) releaseStockForReservation(ctx context.Context, tx *sql.Tx, reservation domain.Reservation) error {
	// Release the reserved stock
	if err := s.productStockRepo.ReleaseStock(ctx, tx, reservation.ProductID, reservation.WarehouseID, int32(reservation.Qty)); err != nil {
		return err
	}

	// Record stock movement for audit trail
	if err := s.movementRepo.Append(ctx, tx, reservation.ProductID, reservation.WarehouseID,
		"RELEASE", reservation.Qty, "RESERVATION_EXPIRED", reservation.ID); err != nil {
		return err
	}

	return nil
}

func (s *stockReleaseUsecase) updateOrderStatusIfAllReservationsExpired(ctx context.Context, tx *sql.Tx, expiredReservations []domain.Reservation) error {
	orderIDs := make(map[string]bool)

	// Collect unique order IDs
	for _, reservation := range expiredReservations {
		orderIDs[reservation.OrderID.String()] = true
	}

	// Check each order to see if all reservations are expired
	for orderIDStr := range orderIDs {
		orderID := expiredReservations[0].OrderID // Get the actual UUID
		for _, res := range expiredReservations {
			if res.OrderID.String() == orderIDStr {
				orderID = res.OrderID
				break
			}
		}

		// Count pending reservations for this order
		pendingCount, err := s.reservationRepo.PendingCountByOrder(ctx, tx, orderID)
		if err != nil {
			return err
		}

		// If no pending reservations left, mark order as expired
		if pendingCount == 0 {
			if err := s.orderRepo.Updatestatus(ctx, orderID, domain.StatusExpired); err != nil {
				return err
			}
			log.Printf("Order %s marked as expired - all reservations have expired", orderID)
		}
	}

	return nil
}

func NewStockReleaseUsecase(
	db pqsql.Database,
	reservationRepo domain.ReservationRepository,
	productStockRepo domain.ProductStockRepository,
	movementRepo domain.MovementRepository,
	orderRepo domain.OrderRepository,
) StockReleaseUsecase {
	return &stockReleaseUsecase{
		db:               db,
		reservationRepo:  reservationRepo,
		productStockRepo: productStockRepo,
		movementRepo:     movementRepo,
		orderRepo:        orderRepo,
	}
}
