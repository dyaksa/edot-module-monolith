package repository

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/google/uuid"
)

type movementRepository struct {
	db pqsql.Client
}

// Append implements domain.MovementRepository.
func (m *movementRepository) Append(ctx context.Context, tx *sql.Tx, productID uuid.UUID, warehouseID uuid.UUID, typ string, qty int, refType string, refID uuid.UUID) error {
	query := squirrel.
		Insert("stock_movements").
		Columns("id", "product_id", "warehouse_id", "type", "qty", "ref_type", "ref_id").
		Values(uuid.New(), productID, warehouseID, typ, qty, refType, refID).
		PlaceholderFormat(squirrel.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)

	return err
}

func NewMovementRepository(db pqsql.Client) domain.MovementRepository {
	return &movementRepository{db: db}
}
