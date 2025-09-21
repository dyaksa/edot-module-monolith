package repository

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/google/uuid"
)

type productRepository struct {
	db pqsql.Client
}

func (p *productRepository) RetrieveAll(ctx context.Context, limit, offset int) ([]domain.RetrieveProduct, error) {
	var results []domain.RetrieveProduct
	avail := sq.Select("s.product_id", "SUM(s.on_hand - s.reserved) AS available", "w.name AS warehouse_name", "w.shop_id AS shop_id").
		From("product_stock s").
		Join("warehouses w ON w.id = s.warehouse_id").
		GroupBy("s.product_id", "w.name", "w.shop_id")

	availSql, availArgs, err := avail.ToSql()
	if err != nil {
		return results, err
	}

	list := sq.Select("p.id", "p.sku", "p.name", "COALESCE(a.available,0) AS available", "a.warehouse_name").
		From("products p").
		LeftJoin("("+availSql+") AS a ON a.product_id = p.id").
		OrderBy("p.created_at DESC", "p.id DESC").
		Limit(uint64(limit)).
		Offset(uint64(offset))

	q, args, err := list.ToSql()
	if err != nil {
		return results, err
	}

	args = append(args, availArgs...)

	rows, err := p.db.Database().QueryContext(ctx, q, args...)
	if err != nil {
		return results, err
	}
	defer rows.Close()

	for rows.Next() {
		var r domain.RetrieveProduct
		if err := rows.Scan(&r.ID, &r.SKU, &r.Name, &r.Available, &r.WarehouseName); err != nil {
			return results, err
		}
		results = append(results, r)
	}
	return results, nil
}

// Create implements domain.ProductRepository.
func (p *productRepository) Create(ctx context.Context, product *domain.Product) (uuid.UUID, error) {
	var id uuid.UUID

	query := sq.Insert("products").
		Columns("sku", "name").
		Values(&product.SKU, &product.Name).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return id, err
	}

	if err := p.db.Database().QueryRowContext(ctx, q, args...).Scan(&id); err != nil {
		fmt.Println("err", err)
		return id, err
	}

	return id, nil
}

func NewProductRepository(db pqsql.Client) domain.ProductRepository {
	return &productRepository{
		db: db,
	}
}
