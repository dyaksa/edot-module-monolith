package paginator

import "context"

type DataFetcher[T any] func(ctx context.Context, offset, limit int) (items []T, totalItems int, err error)

type Paginator[T any] interface {
	Paginate(ctx context.Context, req PaginationRequest, fetcher DataFetcher[T]) (*PaginationResult[T], error)
}
