// Package ptr provides generic pointer utilities for nullable database fields.
package ptr

// To returns a pointer to the given value.
func To[T any](v T) *T { return &v }

// From dereferences a pointer, returning the zero value if nil.
func From[T any](p *T) T {
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
