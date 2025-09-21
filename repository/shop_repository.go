package repository

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/google/uuid"
)

type shopRepository struct {
	db pqsql.Client
}

// Create implements domain.ShopRepository.
func (s *shopRepository) Create(ctx context.Context, shop *domain.Shop) (uuid.UUID, error) {
	var id uuid.UUID

	query := sq.Insert("shops").
		Columns("name", "created_at").
		Values(&shop.Name, time.Now()).
		Suffix("RETURNING id").PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return id, err
	}

	if err := s.db.Database().QueryRowContext(ctx, q, args...).Scan(&id); err != nil {
		return id, err
	}

	return id, nil
}

// Delete implements domain.ShopRepository.
func (s *shopRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := sq.Delete("shops").Where(sq.Eq{"id": id}).PlaceholderFormat(sq.Dollar)
	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.Database().ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}

	return nil
}

// Retrieve implements domain.ShopRepository.
func (s *shopRepository) Retrieve(ctx context.Context, id uuid.UUID) (*domain.Shop, error) {
	var shop domain.Shop

	query := sq.Select("id", "name", "created_at").
		From("shops").
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	if err := s.db.Database().QueryRowContext(ctx, q, args...).Scan(&shop.ID, &shop.Name, &shop.CreatedAt); err != nil {
		return nil, err
	}

	return &shop, nil
}

// Update implements domain.ShopRepository.
func (s *shopRepository) Update(ctx context.Context, shop *domain.Shop) error {
	query := sq.Update("shops").Set("name", shop.Name).Where(sq.Eq{"id": shop.ID}).PlaceholderFormat(sq.Dollar)
	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = s.db.Database().ExecContext(ctx, q, args...)
	if err != nil {
		return err
	}

	return nil
}

func NewShopRepository(db pqsql.Client) domain.ShopRepository {
	return &shopRepository{
		db: db,
	}
}
