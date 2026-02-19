// Package opt provides an optional field type for JSON PATCH semantics.
// It distinguishes three states that plain pointer fields cannot:
//
//   - absent  — field was not sent (do not update)
//   - present — field was sent with a value
//   - null    — field was explicitly set to null (clear the value)
package opt

import "encoding/json"

// Field wraps a value T and tracks whether it was present in the JSON body.
// Use Field[string] for non-nullable fields, Field[*string] for nullable ones.
type Field[T any] struct {
	Value T
	Set   bool
}

// UnmarshalJSON marks the field as set and decodes the value.
func (f *Field[T]) UnmarshalJSON(b []byte) error {
	f.Set = true
	return json.Unmarshal(b, &f.Value)
}
