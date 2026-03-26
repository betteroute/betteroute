// Package page provides pagination utilities for list endpoints.
//
// Uses the LIMIT+1 pattern: fetch one extra row to determine has_more
// without a separate COUNT query. O(limit) instead of O(total).
package page

const (
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// Pagination contains metadata for paginated responses.
type Pagination struct {
	PerPage int  `json:"per_page"`
	HasMore bool `json:"has_more"`
}

// List wraps paginated data with metadata.
type List[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// NewList creates a paginated response from rows fetched with LIMIT+1.
// If len(rows) > perPage, has_more is true and the extra row is trimmed.
// Guarantees data is an empty array [] instead of null when empty.
func NewList[T any](rows []T, perPage int) List[T] {
	hasMore := len(rows) > perPage
	if hasMore {
		rows = rows[:perPage]
	}
	if rows == nil {
		rows = []T{}
	}
	return List[T]{
		Data: rows,
		Pagination: Pagination{
			PerPage: perPage,
			HasMore: hasMore,
		},
	}
}

// NormalizePerPage clamps per_page to safe bounds.
func NormalizePerPage(perPage int) int {
	if perPage < 1 {
		return DefaultPerPage
	}
	if perPage > MaxPerPage {
		return MaxPerPage
	}
	return perPage
}
