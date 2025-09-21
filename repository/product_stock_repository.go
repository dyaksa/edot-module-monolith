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

type productStockRepository struct {
	db pqsql.Client
}

// ReleaseStock implements domain.ProductStockRepository.
func (p *productStockRepository) ReleaseStock(ctx context.Context, tx *sql.Tx, productID uuid.UUID, warehouseID uuid.UUID, quantity int32) error {
	query := sq.Update("product_stock").
		Set("reserved", sq.Expr("reserved - ?", quantity)).
		Set("updated_at", "now()").
		Where(sq.And{
			sq.Eq{"product_id": productID},
			sq.Eq{"warehouse_id": warehouseID},
		}).PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)

	if err != nil {
		return err
	}

	return nil
}

// TryReserveStock implements domain.ProductStockRepository.
func (p *productStockRepository) TryReserveStock(ctx context.Context, tx *sql.Tx, productID uuid.UUID, warehouseID uuid.UUID, quantity int32) (bool, error) {
	query := sq.Update("product_stock").
		Set("reserved", sq.Expr("reserved + ?", quantity)).
		Set("updated_at", "now()").
		Where(sq.And{
			sq.Eq{"product_id": productID},
			sq.Eq{"warehouse_id": warehouseID},
			sq.Expr("(on_hand - reserved) >= ?", quantity),
		}).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return false, err
	}

	var id uuid.UUID
	err = tx.QueryRowContext(ctx, q, args...).Scan(&id)

	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// CommitStock implements domain.ProductStockRepository.
func (p *productStockRepository) CommitStock(ctx context.Context, tx *sql.Tx, productID uuid.UUID, warehouseID uuid.UUID, quantity int32) error {
	query := sq.Update("product_stock").
		Set("on_hand", sq.Expr("on_hand - ?", quantity)).
		Set("reserved", sq.Expr("reserved - ?", quantity)).
		Set("updated_at", "now()").
		Where(sq.And{
			sq.Eq{"product_id": productID},
			sq.Eq{"warehouse_id": warehouseID},
		}).PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)
	return err
}

// AddStock implements domain.ProductStockRepository.
func (p *productStockRepository) AddStock(ctx context.Context, tx *sql.Tx, productID uuid.UUID, warehouseID uuid.UUID, quantity int32) error {
	// First try to update existing record
	updateQuery := sq.Update("product_stock").
		Set("on_hand", sq.Expr("on_hand + ?", quantity)).
		Set("updated_at", "now()").
		Where(sq.And{
			sq.Eq{"product_id": productID},
			sq.Eq{"warehouse_id": warehouseID},
		}).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	q, args, err := updateQuery.ToSql()
	if err != nil {
		return err
	}

	var id uuid.UUID
	err = tx.QueryRowContext(ctx, q, args...).Scan(&id)

	if err == sql.ErrNoRows {
		// Product doesn't exist in this warehouse, create new record
		insertQuery := sq.Insert("product_stock").
			Columns("id", "product_id", "warehouse_id", "on_hand", "reserved", "updated_at").
			Values(uuid.New(), productID, warehouseID, quantity, 0, "now()").
			PlaceholderFormat(sq.Dollar)

		iq, iargs, err := insertQuery.ToSql()
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, iq, iargs...)
		return err
	}

	return err
}

func (p *productStockRepository) Create(ctx context.Context, productStock *domain.ProductStock) (uuid.UUID, error) {
	var id uuid.UUID
	query := sq.Insert("product_stock").
		Columns("product_id", "warehouse_id", "on_hand", "updated_at").
		Values(&productStock.ProductID, &productStock.WarehouseID, &productStock.OnHand, time.Now()).
		Suffix("RETURNING id").PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return id, err
	}

	if err := p.db.Database().QueryRowContext(ctx, q, args...).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}

func NewProductStockRepository(db pqsql.Client) domain.ProductStockRepository {
	return &productStockRepository{
		db: db,
	}
}
