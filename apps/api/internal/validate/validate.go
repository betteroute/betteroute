// Package validate provides a shared struct validator backed by
// go-playground/validator. Handlers call Struct to validate input
// and receive []errs.FieldError for structured 422 responses.
package validate

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"github.com/execrc/betteroute/internal/errs"
	"github.com/execrc/betteroute/internal/opt"
)

// v is the shared validator instance.
var v = validator.New(validator.WithRequiredStructEnabled())

func init() {
	// Use JSON tag names in error messages so field names match the API contract.
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name, _, _ := strings.Cut(fld.Tag.Get("json"), ",")
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})

	_ = v.RegisterValidation("shortcode", isShortCode)

	// Unwrap opt.Field[T] so validator sees the inner Value for tag-based checks.
	v.RegisterCustomTypeFunc(extractOptField,
		opt.Field[string]{},
		opt.Field[*string]{},
		opt.Field[int32]{},
		opt.Field[*int32]{},
		opt.Field[bool]{},
		opt.Field[*bool]{},
		opt.Field[*time.Time]{},
	)
}

// extractOptField returns the Value inside an opt.Field so the validator
// can apply struct tags (omitempty, url, max, etc.) to the actual value.
func extractOptField(field reflect.Value) any {
	return field.FieldByName("Value").Interface()
}

// Struct validates the given struct and returns field errors, if any.
// Returns nil when validation passes.
func Struct(s any) []errs.FieldError {
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return []errs.FieldError{{Field: "_", Message: "invalid input"}}
	}

	fields := make([]errs.FieldError, 0, len(validationErrors))
	for _, fe := range validationErrors {
		fields = append(fields, errs.FieldError{
			Field:   fe.Field(),
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

var shortCodeRe = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// isShortCode validates that s contains only [a-zA-Z0-9_-].
func isShortCode(fl validator.FieldLevel) bool {
	return shortCodeRe.MatchString(fl.Field().String())
}
