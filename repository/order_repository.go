package repository

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/google/uuid"
)

type orderRepository struct {
	db pqsql.Client
}

// Create implements domain.OrderRepository.
func (or *orderRepository) Create(ctx context.Context, tx *sql.Tx, o *domain.Order) error {
	query := sq.Insert("orders").
		Columns("id", "user_id", "shop_id", "status", "total_amount").
		Values(o.ID, o.UserID, o.ShopID, o.Status, o.Total).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)

	return err
}

// GetByID implements domain.OrderRepository.
func (or *orderRepository) GetByID(ctx context.Context, orderID uuid.UUID) (*domain.Order, error) {
	query := sq.Select("id", "user_id", "shop_id", "status", "CAST(total_amount AS BIGINT) as total_amount", "created_at", "updated_at").
		From("orders").
		Where(sq.Eq{"id": orderID}).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var o domain.Order
	err = or.db.Database().QueryRowContext(ctx, q, args...).
		Scan(&o.ID, &o.UserID, &o.ShopID, &o.Status, &o.Total, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return &o, nil
}

// Updatestatus implements domain.OrderRepository.
func (or *orderRepository) Updatestatus(ctx context.Context, orderID uuid.UUID, status domain.OrderStatus) error {
	query := sq.Update("orders").
		Set("status", status).
		Set("updated_at", sq.Expr("now()")).
		Where(sq.Eq{"id": orderID}).
		PlaceholderFormat(sq.Dollar)
	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = or.db.Database().ExecContext(ctx, q, args...)

	return err
}

// GetByUserID implements domain.OrderRepository.
func (or *orderRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.OrderListItem, int, error) {
	var orders []domain.OrderListItem
	var totalCount int

	// First, get the total count for pagination
	countQuery := sq.Select("COUNT(*)").
		From("orders").
		Where(sq.Eq{"user_id": userID}).
		PlaceholderFormat(sq.Dollar)

	countSql, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, 0, err
	}

	err = or.db.Database().QueryRowContext(ctx, countSql, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Then get the paginated orders with item count
	query := sq.Select(
		"o.id",
		"CAST(o.total_amount AS BIGINT) as total_amount",
		"o.status",
		"COUNT(oi.id) as item_count",
		"o.reservation_expires_at",
		"o.created_at",
	).
		From("orders o").
		LeftJoin("order_items oi ON oi.order_id = o.id").
		Where(sq.Eq{"o.user_id": userID}).
		GroupBy("o.id", "o.total_amount", "o.status", "o.reservation_expires_at", "o.created_at").
		OrderBy("o.created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset)).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return nil, 0, err
	}

	rows, err := or.db.Database().QueryContext(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var order domain.OrderListItem
		var reservationExpiresAt sql.NullTime

		err := rows.Scan(
			&order.ID,
			&order.Total,
			&order.Status,
			&order.ItemCount,
			&reservationExpiresAt,
			&order.CreatedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if reservationExpiresAt.Valid {
			order.ReservationExpiresAt = &reservationExpiresAt.Time
		}

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return orders, totalCount, nil
}

func NewOrderRepository(db pqsql.Client) domain.OrderRepository {
	return &orderRepository{db: db}
}
