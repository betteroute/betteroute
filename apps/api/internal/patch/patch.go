// Package patch provides helpers for PATCH request handling.
// Bind[T] parses a JSON body while tracking which top-level keys
// were present, enabling PATCH semantics where "field omitted"
// (don't change) differs from "field: null" (clear the value).
package patch

import "encoding/json"

// Fields tracks which JSON keys were present in a request body.
type Fields map[string]struct{}

// Has reports whether the given JSON key was present.
func (f Fields) Has(key string) bool {
	_, ok := f[key]
	return ok
}

// Bind unmarshals a JSON body into T and returns the parsed value
// along with the set of top-level keys that were present.
// This allows PATCH handlers to distinguish absent fields from
// fields explicitly set to null.
func Bind[T any](data []byte) (T, Fields, error) {
	var zero T

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return zero, nil, err
	}

	fields := make(Fields, len(raw))
	for k := range raw {
		fields[k] = struct{}{}
	}

	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return zero, nil, err
	}

	return v, fields, nil
}
