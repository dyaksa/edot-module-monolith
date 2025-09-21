package repository

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/infrastructure/pqsql"
	"github.com/google/uuid"
)

type idempotencyRequestRepository struct {
	db pqsql.Client
}

// BeginKey implements domain.IdempotencyRequestRepository.
func (i *idempotencyRequestRepository) BeginKey(ctx context.Context, tx *sql.Tx, key string, endpoint string, payloadHash string) (isNew bool, err error) {
	query := sq.Insert("idempotency_requests").
		Columns("key", "endpoint", "payload_hash").Values(key, endpoint, payloadHash).
		Suffix("ON CONFLICT (key, endpoint) DO NOTHING").
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return false, err
	}

	res, err := tx.ExecContext(ctx, q, args...)
	if err != nil {
		return false, err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return rowsAffected > 0, nil
}

// LoadIfExists implements domain.IdempotencyRequestRepository.
func (i *idempotencyRequestRepository) LoadIfExists(ctx context.Context, tx *sql.Tx, key string, endpoint string) (payloadHash string, orderID *uuid.UUID, responseJSON []byte, exists bool, err error) {
	query := sq.Select("payload_hash", "order_id", "response_body").
		From("idempotency_requests").
		Where(sq.And{
			sq.Eq{"key": key},
			sq.Eq{"endpoint": endpoint},
		}).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return "", nil, nil, false, err
	}

	var oID sql.NullString
	var respJSON []byte
	err = tx.QueryRowContext(ctx, q, args...).Scan(&payloadHash, &oID, &respJSON)

	switch {
	case err == sql.ErrNoRows:
		return "", nil, nil, false, nil
	case err != nil:
		return "", nil, nil, false, err
	}

	if oID.Valid {
		parsedID, err := uuid.Parse(oID.String)
		if err != nil {
			return "", nil, nil, false, err
		}
		orderID = &parsedID
	}

	responseJSON = respJSON

	return payloadHash, orderID, responseJSON, true, nil
}

// SaveResponse implements domain.IdempotencyRequestRepository.
func (i *idempotencyRequestRepository) SaveResponse(ctx context.Context, tx *sql.Tx, key string, endpoint string, orderID uuid.UUID, responseJSON []byte) error {
	query := sq.Update("idempotency_requests").
		Set("order_id", orderID).
		Set("response_body", responseJSON).
		Where(sq.And{
			sq.Eq{"key": key},
			sq.Eq{"endpoint": endpoint},
		}).
		PlaceholderFormat(sq.Dollar)

	q, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, q, args...)

	return err
}

func NewIdempotencyRequestRepository(db pqsql.Client) domain.IdempotencyRequestRepository {
	return &idempotencyRequestRepository{db: db}
}
