package usecase

import (
	"context"

	"github.com/dyaksa/warehouse/domain"
	"github.com/dyaksa/warehouse/pkg/errx"
	"github.com/dyaksa/warehouse/pkg/paginator"
	"github.com/google/uuid"
)

type productUsecase struct {
	productRepository      domain.ProductRepository
	productStockRepository domain.ProductStockRepository
	paginator              paginator.Paginator[domain.RetrieveProduct]
}

// RetrieveAll implements domain.ProductUsecase.
func (pu *productUsecase) RetrieveAll(ctx context.Context, pagination paginator.PaginationRequest) (*paginator.PaginationResult[domain.RetrieveProduct], error) {
	return pu.paginator.Paginate(ctx, pagination, func(ctx context.Context, offset, limit int) (items []domain.RetrieveProduct, totalItems int, err error) {
		items, err = pu.productRepository.RetrieveAll(ctx, limit, offset)
		if err != nil {
			return items, totalItems, errx.E(errx.CodeInternal, "failed to retrieve products", errx.Op("productUsecase.RetrieveAll"), err)
		}

		totalItems = len(items)
		return items, totalItems, nil
	})
}

func (pu *productUsecase) Create(ctx context.Context, payload domain.CreateProductRequest) error {
	warehouseId, err := uuid.Parse(payload.WarehouseID)
	if err != nil {
		return errx.E(errx.CodeValidation, "invalid warehouse UUID", errx.Op("productUsecase.Create"), err)
	}

	product := domain.Product{
		SKU:  payload.SKU,
		Name: payload.Name,
	}

	productId, err := pu.productRepository.Create(ctx, &product)
	if err != nil {
		return errx.E(errx.CodeInternal, "failed to create product", errx.Op("productUsecase.Create"), err)
	}

	productStock := domain.ProductStock{
		WarehouseID: warehouseId,
		ProductID:   productId,
		OnHand:      payload.OnHand,
	}

	if _, err = pu.productStockRepository.Create(ctx, &productStock); err != nil {
		return errx.E(errx.CodeInternal, "failed to create product stock", errx.Op("productUsecase.Create"), err)
	}

	return nil
}

func NewProductUsecase(
	productRepository domain.ProductRepository,
	productStockUsecase domain.ProductStockRepository,
) domain.ProductUsecase {
	return &productUsecase{
		productRepository:      productRepository,
		productStockRepository: productStockUsecase,
		paginator:              paginator.NewOffsetPaginator[domain.RetrieveProduct](),
	}
}
