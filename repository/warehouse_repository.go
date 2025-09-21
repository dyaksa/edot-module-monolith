package repository

import (
	"context"
	"database/sql"
	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/google/uuid"
)

type warehouseRepository struct {
	db pqsql.Client
}

// Pick implements domain.WarehouseRepository.
func (wr *warehouseRepository) Pick(ctx context.Context, tx *sql.Tx, productID uuid.UUID, qty int, shopID uuid.UUID) (warehouseID uuid.UUID, err error) {
	query := sq.Select("ps.warehouse_id").
		From("product_stock ps").
		Join("warehouses w ON w.id = ps.warehouse_id").
		Where(sq.And{
			sq.Eq{"ps.product_id": productID},
			sq.Eq{"w.shop_id": shopID},
			sq.Expr("(ps.on_hand - ps.reserved) >= ?", qty),
			sq.Eq{"w.is_active": true},
		}).
		GroupBy("ps.warehouse_id", "ps.product_id", "w.is_active").
		OrderBy("SUM(ps.on_hand - ps.reserved) DESC", "ps.warehouse_id ASC").
		Limit(1).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return uuid.Nil, err
	}

	err = tx.QueryRowContext(ctx, q, args...).Scan(&warehouseID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, domain.ErrOutOfStock
		}
		return uuid.Nil, err
	}

	return warehouseID, nil
}

// Delete implements domain.WarehouseRepository.
func (wr *warehouseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := wr.db.Database().ExecContext(ctx, "DELETE FROM warehouses WHERE id = $1", id)
	if err != nil {
		return err
	}

	return nil
}

// Retrieve implements domain.WarehouseRepository.
func (wr *warehouseRepository) Retrieve(ctx context.Context, id uuid.UUID) (*domain.WareHouse, error) {
	var w domain.WareHouse
	err := wr.db.Database().QueryRowContext(ctx, "SELECT id, shop_id, name, is_active, created_at FROM warehouses WHERE id = $1", id).Scan(&w.ID, &w.ShopID, &w.Name, &w.IsActive, &w.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

// Update implements domain.WarehouseRepository.
func (wr *warehouseRepository) Update(ctx context.Context, w *domain.WareHouse) error {
	_, err := wr.db.Database().ExecContext(ctx, "UPDATE warehouses SET shop_id = $2, name = $3, is_active = $4 WHERE id = $1", w.ID, w.ShopID, w.Name, w.IsActive)
	if err != nil {
		return err
	}

	return nil
}

// Create implements domain.WarehouseRepository.
func (wr *warehouseRepository) Create(ctx context.Context, w *domain.WareHouse) error {
	_, err := wr.db.Database().Transaction(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		query := `INSERT INTO warehouses (shop_id, name, is_active, created_at) VALUES ($1, $2, $3, now())`
		_, err := tx.ExecContext(ctx, query, w.ShopID, w.Name, w.IsActive)
		return nil, err
	})

	if err != nil {
		return err
	}

	return nil
}

// GetByShopID implements domain.WarehouseRepository.
func (wr *warehouseRepository) GetByShopID(ctx context.Context, shopID uuid.UUID) ([]domain.WareHouse, error) {
	query := `SELECT id, shop_id, name, is_active, created_at FROM warehouses WHERE shop_id = $1 ORDER BY created_at DESC`
	rows, err := wr.db.Database().QueryContext(ctx, query, shopID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var warehouses []domain.WareHouse
	for rows.Next() {
		var w domain.WareHouse
		err := rows.Scan(&w.ID, &w.ShopID, &w.Name, &w.IsActive, &w.CreatedAt)
		if err != nil {
			return nil, err
		}
		warehouses = append(warehouses, w)
	}

	return warehouses, nil
}

// SetActive implements domain.WarehouseRepository.
func (wr *warehouseRepository) SetActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	_, err := wr.db.Database().ExecContext(ctx, "UPDATE warehouses SET is_active = $2 WHERE id = $1", id, isActive)
	return err
}

func NewWarehouseRepository(db pqsql.Client) domain.WarehouseRepository {
	return &warehouseRepository{db: db}
}
