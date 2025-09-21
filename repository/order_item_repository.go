package repository

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
)

type orderItemRepository struct {
	db pqsql.Client
}

func (o *orderItemRepository) BulkInsert(ctx context.Context, tx *sql.Tx, items []domain.OrderItem) error {
	if len(items) == 0 {
		return nil
	}

	query := sq.Insert("order_items").
		Columns("id", "order_id", "product_id", "qty", "price").
		PlaceholderFormat(sq.Dollar)

	for _, item := range items {
		query = query.Values(item.ID, item.OrderID, item.ProductID, item.Qty, item.Price)
	}

	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)
	return err
}

func NewOrderItemRepository(db pqsql.Client) domain.OrderItemRepository {
	return &orderItemRepository{db: db}
}
