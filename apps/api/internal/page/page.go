// Package page provides pagination utilities for list endpoints.
package page

// Pagination defaults and limits.
const (
	DefaultPage    = 1
	DefaultPerPage = 20
	MaxPerPage     = 100
)

// Pagination contains metadata for paginated responses.
type Pagination struct {
	Page       int  `json:"page"`
	PerPage    int  `json:"per_page"`
	Total      int  `json:"total"`
	TotalPages int  `json:"total_pages"`
	HasMore    bool `json:"has_more"`
}

// List wraps paginated data with metadata.
type List[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// NewList creates a paginated list response.
// Guarantees data is an empty array [] instead of null when empty.
func NewList[T any](data []T, pg, perPg, total int) List[T] {
	if data == nil {
		data = []T{}
	}

	totalPages := 0
	if perPg > 0 {
		totalPages = (total + perPg - 1) / perPg
	}

	return List[T]{
		Data: data,
		Pagination: Pagination{
			Page:       pg,
			PerPage:    perPg,
			Total:      total,
			TotalPages: totalPages,
			HasMore:    pg < totalPages,
		},
	}
}

// Normalize clamps page and perPage to safe values.
func Normalize(pg, perPg int) (int, int) {
	if pg < 1 {
		pg = DefaultPage
	}
	if perPg < 1 {
		perPg = DefaultPerPage
	}
	if perPg > MaxPerPage {
		perPg = MaxPerPage
	}
	return pg, perPg
}

// Offset calculates SQL OFFSET from page and perPage.
func Offset(pg, perPg int) int {
	return (pg - 1) * perPg
}
