package paginator

import "context"

type OffsetPaginator[T any] struct{}

func (o *OffsetPaginator[T]) Paginate(ctx context.Context, req PaginationRequest, fetcher DataFetcher[T]) (*PaginationResult[T], error) {
	req.ValidateAndSetDefault()
	offset := req.GetOffset()

	items, totalItems, err := fetcher(ctx, offset, req.Limit)
	if err != nil {
		return nil, err
	}

	return NewPaginationResult(items, totalItems, req), nil
}

func NewOffsetPaginator[T any]() Paginator[T] {
	return &OffsetPaginator[T]{}
}
