package domain

import (
	"context"

	"github.com/google/uuid"
)

type Shop struct {
	ID        uuid.UUID
	Name      string
	CreatedAt string
}

type CreateShopRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateShopRequest struct {
	ID   uuid.UUID `form:"id" binding:"required"`
	Name string    `json:"name" binding:"required"`
}

type ShopQuery struct {
	ID uuid.UUID `form:"id" binding:"required"`
}

type ShopRepository interface {
	Create(ctx context.Context, shop *Shop) (uuid.UUID, error)
	Retrieve(ctx context.Context, id uuid.UUID) (*Shop, error)
	Update(ctx context.Context, shop *Shop) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type ShopUsecase interface {
	Create(ctx context.Context, payload CreateShopRequest) error
	Retrieve(ctx context.Context, id uuid.UUID) (*Shop, error)
	Update(ctx context.Context, payload UpdateShopRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
}
