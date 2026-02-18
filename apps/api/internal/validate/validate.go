// Package validate provides a shared struct validator backed by
// go-playground/validator. Handlers call Struct to validate input
// and receive []errs.FieldError for structured 422 responses.
package validate

import (
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/execrc/betteroute/internal/errs"
)

// v is the shared validator instance.
var v = validator.New(validator.WithRequiredStructEnabled())

func init() {
	v.RegisterValidation("shortcode", isShortCode)
}

// Struct validates the given struct and returns field errors, if any.
// Returns nil when validation passes.
func Struct(s any) []errs.FieldError {
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []errs.FieldError{{Field: "_", Message: "invalid input"}}
	}

	fields := make([]errs.FieldError, 0, len(validationErrors))
	for _, fe := range validationErrors {
		fields = append(fields, errs.FieldError{
			Field:   toSnakeCase(fe.Field()),
			Message: message(fe),
		})
	}
	return fields
}

// message returns a human-readable validation message.
func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "url":
		return "must be a valid URL"
	case "min":
		return "must be at least " + fe.Param() + " characters"
	case "max":
		return "must be at most " + fe.Param() + " characters"
	case "oneof":
		return "must be one of: " + fe.Param()
	case "email":
		return "must be a valid email address"
	case "gt":
		return "must be greater than " + fe.Param()
	case "gte":
		return "must be at least " + fe.Param()
	case "hexcolor":
		return "must be a valid hex color (#RRGGBB)"
	case "len":
		return "must be exactly " + fe.Param() + " characters"
	case "shortcode":
		return "must contain only letters, numbers, hyphens, and underscores"
	default:
		return "failed " + fe.Tag() + " validation"
	}
}

// isShortCode validates that s contains only [a-zA-Z0-9_-].
func isShortCode(fl validator.FieldLevel) bool {
	s := fl.Field().String()
	if len(s) == 0 {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// toSnakeCase converts PascalCase field names to snake_case.
// "DestURL" → "dest_url", "WorkspaceID" → "workspace_id".
func toSnakeCase(s string) string {
	var b strings.Builder
	b.Grow(len(s) + 4)

	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				prev := s[i-1]
				if prev >= 'a' && prev <= 'z' {
					b.WriteByte('_')
				} else if i+1 < len(s) && s[i+1] >= 'a' && s[i+1] <= 'z' {
					b.WriteByte('_')
				}
			}
			b.WriteRune(r + 32)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}
