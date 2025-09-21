package repository

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/google/uuid"
)

type reservationRepository struct {
	db pqsql.Client
}

// CreateMany implements domain.ReservationRepository.
func (r *reservationRepository) CreateMany(ctx context.Context, tx *sql.Tx, reservations []domain.Reservation) error {
	query := sq.Insert("stock_reservations").
		Columns("id", "order_id", "product_id", "warehouse_id", "qty", "status", "expires_at").
		PlaceholderFormat(sq.Dollar)

	for _, res := range reservations {
		query = query.Values(res.ID, res.OrderID, res.ProductID, res.WarehouseID, res.Qty, res.Status, res.ExpiresAt)
	}

	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)

	return err
}

// MarkExpired implements domain.ReservationRepository.
func (r *reservationRepository) MarkExpired(ctx context.Context, tx *sql.Tx, id uuid.UUID) error {
	query := sq.Update("stock_reservations").
		Set("status", "EXPIRED").
		Set("updated_at", time.Now()).
		Where(sq.And{
			sq.Eq{"id": id},
			sq.Eq{"status": "PENDING"},
		}).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)

	return err
}

// PendingCountByOrder implements domain.ReservationRepository.
func (r *reservationRepository) PendingCountByOrder(ctx context.Context, tx *sql.Tx, orderID uuid.UUID) (int, error) {
	var n int
	query := sq.Select("COUNT(*)").
		From("stock_reservations").
		Where(sq.And{
			sq.Eq{"order_id": orderID},
			sq.Eq{"status": "PENDING"},
		}).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	err = tx.QueryRowContext(ctx, q, args...).Scan(&n)
	if err != nil {
		return 0, err
	}

	return n, nil
}

// PickExpiredForUpdate implements domain.ReservationRepository.
func (r *reservationRepository) PickExpiredForUpdate(ctx context.Context, tx *sql.Tx, limit int) ([]domain.Reservation, error) {
	query := sq.Select("id", "order_id", "product_id", "warehouse_id", "qty", "status", "expires_at").
		From("stock_reservations").
		Where(sq.And{
			sq.Eq{"status": "PENDING"},
			sq.Expr("expires_at <= now()"),
		}).
		Limit(uint64(limit)).
		Suffix("FOR UPDATE SKIP LOCKED").
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []domain.Reservation
	for rows.Next() {
		var res domain.Reservation
		err := rows.Scan(&res.ID, &res.OrderID, &res.ProductID, &res.WarehouseID,
			&res.Qty, &res.Status, &res.ExpiresAt)

		if err != nil {
			return nil, err
		}

		reservations = append(reservations, res)
	}

	return reservations, rows.Err()
}

// Retrieve implements domain.ReservationRepository.
func (r *reservationRepository) Retrieve(ctx context.Context, tx *sql.Tx, id uuid.UUID) (*domain.Reservation, error) {
	var reservation domain.Reservation
	query := sq.Select("id", "order_id", "product_id", "warehouse_id", "qty", "status", "expires_at").
		From("stock_reservations").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	err = tx.QueryRowContext(ctx, q, args...).
		Scan(&reservation.ID, &reservation.OrderID, &reservation.ProductID, &reservation.WarehouseID,
			&reservation.Qty, &reservation.Status, &reservation.ExpiresAt)
	if err != nil {
		return nil, err
	}

	return &reservation, nil
}

// MarkCommitted implements domain.ReservationRepository.
func (r *reservationRepository) MarkCommitted(ctx context.Context, tx *sql.Tx, orderID uuid.UUID) error {
	query := sq.Update("stock_reservations").
		Set("status", "COMMITTED").
		Set("updated_at", time.Now()).
		Where(sq.And{
			sq.Eq{"order_id": orderID},
			sq.Eq{"status": "PENDING"},
		}).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)
	return err
}

// MarkReleased implements domain.ReservationRepository.
func (r *reservationRepository) MarkReleased(ctx context.Context, tx *sql.Tx, orderID uuid.UUID) error {
	query := sq.Update("stock_reservations").
		Set("status", "RELEASED").
		Set("updated_at", time.Now()).
		Where(sq.And{
			sq.Eq{"order_id": orderID},
			sq.Eq{"status": "PENDING"},
		}).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)
	return err
}

// GetByOrderID implements domain.ReservationRepository.
func (r *reservationRepository) GetByOrderID(ctx context.Context, tx *sql.Tx, orderID uuid.UUID) ([]domain.Reservation, error) {
	query := sq.Select("id", "order_id", "product_id", "warehouse_id", "qty", "status", "expires_at").
		From("stock_reservations").
		Where(sq.Eq{"order_id": orderID}).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reservations []domain.Reservation
	for rows.Next() {
		var res domain.Reservation
		err := rows.Scan(&res.ID, &res.OrderID, &res.ProductID, &res.WarehouseID,
			&res.Qty, &res.Status, &res.ExpiresAt)
		if err != nil {
			return nil, err
		}
		reservations = append(reservations, res)
	}

	return reservations, rows.Err()
}

func NewReservationRepository(db pqsql.Client) domain.ReservationRepository {
	return &reservationRepository{
		db: db,
	}
}
