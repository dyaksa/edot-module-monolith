package domain

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

type IdempotencyRequestRepository interface {
	BeginKey(ctx context.Context, tx *sql.Tx, key, endpoint, payloadHash string) (isNew bool, err error)
	LoadIfExists(ctx context.Context, tx *sql.Tx, key, endpoint string) (payloadHash string, orderID *uuid.UUID, responseJSON []byte, exists bool, err error)
	SaveResponse(ctx context.Context, tx *sql.Tx, key, endpoint string, orderID uuid.UUID, responseJSON []byte) error
}
