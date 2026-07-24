package dx

import "math"

// PageRequest encapsulates page number and items per page.
type PageRequest struct {
	Page    int `json:"page"`    // 1-indexed page number
	PerPage int `json:"perPage"` // Items per page
}

// PageResult represents a paginated subset of items with metadata.
type PageResult[T any] struct {
	Items      []T   `json:"items"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"perPage"`
	TotalPages int   `json:"totalPages"`
	HasNext    bool  `json:"hasNext"`
	HasPrev    bool  `json:"hasPrev"`
}

// PaginateSlice returns a PageResult for an in-memory slice of items.
func PaginateSlice[T any](items []T, req PageRequest) PageResult[T] {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	perPage := req.PerPage
	if perPage <= 0 {
		perPage = 20
	}

	total := int64(len(items))
	totalPages := int(math.Ceil(float64(total) / float64(perPage)))
	if totalPages <= 0 {
		totalPages = 1
	}

	start := (page - 1) * perPage
	if start > len(items) {
		start = len(items)
	}
	end := start + perPage
	if end > len(items) {
		end = len(items)
	}

	slicedItems := make([]T, 0)
	if start < end {
		slicedItems = items[start:end]
	}

	return PageResult[T]{
		Items:      slicedItems,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
