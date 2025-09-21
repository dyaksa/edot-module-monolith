package usecase

import (
	"context"

	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/pkg/errx"
	"github.com/google/uuid"
)

type shopUsecase struct {
	shopRepository domain.ShopRepository
}

// Create implements domain.ShopUsecase.
func (s *shopUsecase) Create(ctx context.Context, payload domain.CreateShopRequest) error {
	shop := &domain.Shop{
		Name: payload.Name,
	}

	if _, err := s.shopRepository.Create(ctx, shop); err != nil {
		return errx.E(errx.CodeInternal, "failed to create shop", errx.Op("shopUsecase.Create"), err)
	}

	return nil
}

// Delete implements domain.ShopUsecase.
func (s *shopUsecase) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.shopRepository.Delete(ctx, id); err != nil {
		return errx.E(errx.CodeInternal, "failed to delete shop", errx.Op("shopUsecase.Delete"), err)
	}

	return nil
}

// Retrieve implements domain.ShopUsecase.
func (s *shopUsecase) Retrieve(ctx context.Context, id uuid.UUID) (*domain.Shop, error) {
	shop, err := s.shopRepository.Retrieve(ctx, id)
	if err != nil {
		return nil, errx.E(errx.CodeInternal, "failed to retrieve shop", errx.Op("shopUsecase.Retrieve"), err)
	}

	return shop, nil
}

// Update implements domain.ShopUsecase.
func (s *shopUsecase) Update(ctx context.Context, payload domain.UpdateShopRequest) error {
	shop := &domain.Shop{
		ID:   payload.ID,
		Name: payload.Name,
	}

	if err := s.shopRepository.Update(ctx, shop); err != nil {
		return errx.E(errx.CodeInternal, "failed to update shop", errx.Op("shopUsecase.Update"), err)
	}

	return nil
}

func NewShopUsecase(shopRepository domain.ShopRepository) domain.ShopUsecase {
	return &shopUsecase{
		shopRepository: shopRepository,
	}
}
