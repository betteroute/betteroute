// Package ptr provides generic pointer utilities for nullable database fields.
package ptr

import "math"

// Val dereferences a pointer, returning the zero value if nil.
func Val[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}

// ToNonZero returns a pointer to the value, or nil if it equals its zero value.
// Use for DB fields that should store NULL instead of empty strings or zeros.
func ToNonZero[T comparable](v T) *T {
	var zero T
	if v == zero {
		return nil
	}
	return &v
}

// ToInt32 safely converts an int to an int32, clamping if it overflows.
// Useful for dealing with sqlc-generated types without triggering gosec G115.
func ToInt32(v int) int32 {
	if v > math.MaxInt32 {
		return math.MaxInt32
	}
	if v < math.MinInt32 {
		return math.MinInt32
	}
	return int32(v)
}
