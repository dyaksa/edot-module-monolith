package paginator

import "math"

const (
	DefaultPage  = 1
	DefaultLimit = 10
	MaxLimit     = 100
)

type PaginationRequest struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}

func (pr *PaginationRequest) ValidateAndSetDefault() {
	if pr.Page <= 0 {
		pr.Page = DefaultPage
	}

	if pr.Limit <= 0 {
		pr.Limit = DefaultLimit
	}

	if pr.Limit > MaxLimit {
		pr.Limit = MaxLimit
	}
}

func (pr *PaginationRequest) GetOffset() int {
	return (pr.Page - 1) * pr.Limit
}

type PaginationResult[T any] struct {
	Items      []T  `json:"items"`
	TotalItems int  `json:"total_items"`
	TotalPages int  `json:"total_pages"`
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	HasNext    bool `json:"has_next"`
	HasPrev    bool `json:"has_prev"`
}

func NewPaginationResult[T any](items []T, totalItems int, req PaginationRequest) *PaginationResult[T] {
	if items == nil {
		items = make([]T, 0)
	}

	totalPages := 0
	if totalItems > 0 && req.Limit > 0 {
		totalPages = int(math.Ceil(float64(totalItems) / float64(req.Limit)))
	}

	return &PaginationResult[T]{
		Items:      items,
		TotalItems: totalItems,
		TotalPages: totalPages,
		Page:       req.Page,
		Limit:      req.Limit,
		HasNext:    req.Page < totalPages,
		HasPrev:    req.Page > 1 && totalItems > 0,
	}
}
