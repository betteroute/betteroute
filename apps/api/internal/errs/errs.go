// Package errs provides RFC 9457 Problem Details for HTTP API error responses.
// Handlers return *Error values; the Fiber ErrorHandler serializes them as
// application/problem+json with the appropriate status code.
package errs

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

const (
	baseURI     = "https://api.betteroute.dev/errors"
	contentType = "application/problem+json"
)

// FieldError represents a validation error for a specific input field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error is an RFC 9457 Problem Details response.
type Error struct {
	Type       string       `json:"type"`
	Title      string       `json:"title"`
	Status     int          `json:"status"`
	Detail     string       `json:"detail,omitempty"`
	Instance   string       `json:"instance,omitempty"`
	RequestID  string       `json:"request_id,omitempty"`
	Errors     []FieldError `json:"errors,omitempty"`
	RetryAfter int          `json:"retry_after,omitempty"`
	Cause      error        `json:"-"`
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.Detail != "" {
		return e.Title + ": " + e.Detail
	}
	return e.Title
}

// Unwrap returns the underlying cause for errors.Is/As compatibility.
func (e *Error) Unwrap() error { return e.Cause }

// WithCause attaches an underlying error (logged server-side, never sent to client).
func (e *Error) WithCause(err error) *Error {
	if err == nil {
		return e
	}
	cp := *e
	cp.Cause = err
	return &cp
}

// LogValue implements slog.LogValuer for structured logging.
func (e *Error) LogValue() slog.Value {
	attrs := []slog.Attr{slog.Int("status", e.Status)}
	if e.Detail != "" {
		attrs = append(attrs, slog.String("detail", e.Detail))
	}
	if e.Cause != nil {
		attrs = append(attrs, slog.String("cause", e.Cause.Error()))
	}
	return slog.GroupValue(attrs...)
}

// --- Constructors ---

// BadRequest creates a 400 error.
func BadRequest(detail string) *Error {
	return &Error{
		Type:   baseURI + "/bad-request",
		Title:  "Bad Request",
		Status: http.StatusBadRequest,
		Detail: detail,
	}
}

// Unauthorized creates a 401 error.
func Unauthorized(detail string) *Error {
	if detail == "" {
		detail = "Valid authentication credentials are required"
	}
	return &Error{
		Type:   baseURI + "/unauthorized",
		Title:  "Unauthorized",
		Status: http.StatusUnauthorized,
		Detail: detail,
	}
}

// PaymentRequired creates a 402 error.
func PaymentRequired(detail string) *Error {
	if detail == "" {
		detail = "Payment required to access this resource or feature"
	}
	return &Error{
		Type:   baseURI + "/payment-required",
		Title:  "Payment Required",
		Status: http.StatusPaymentRequired,
		Detail: detail,
	}
}

// Forbidden creates a 403 error.
func Forbidden(detail string) *Error {
	if detail == "" {
		detail = "You don't have permission to access this resource"
	}
	return &Error{
		Type:   baseURI + "/forbidden",
		Title:  "Forbidden",
		Status: http.StatusForbidden,
		Detail: detail,
	}
}

// NotFound creates a 404 error.
func NotFound(resource, id string) *Error {
	detail := resource + " not found"
	if id != "" {
		detail = resource + " '" + id + "' not found"
	}
	return &Error{
		Type:   baseURI + "/not-found",
		Title:  "Not Found",
		Status: http.StatusNotFound,
		Detail: detail,
	}
}

// Conflict creates a 409 error.
func Conflict(detail string) *Error {
	return &Error{
		Type:   baseURI + "/conflict",
		Title:  "Conflict",
		Status: http.StatusConflict,
		Detail: detail,
	}
}

// Validation creates a 422 error with per-field details.
func Validation(fields []FieldError) *Error {
	return &Error{
		Type:   baseURI + "/validation-failed",
		Title:  "Validation Failed",
		Status: http.StatusUnprocessableEntity,
		Detail: "One or more fields failed validation",
		Errors: fields,
	}
}

// TooManyRequests creates a 429 error with a Retry-After hint.
func TooManyRequests(retryAfter int) *Error {
	return &Error{
		Type:       baseURI + "/rate-limited",
		Title:      "Too Many Requests",
		Status:     http.StatusTooManyRequests,
		Detail:     "Rate limit exceeded",
		RetryAfter: retryAfter,
	}
}

// Internal creates a 500 error. The detail is never empty to avoid blank responses.
func Internal(detail string) *Error {
	if detail == "" {
		detail = "An unexpected error occurred"
	}
	return &Error{
		Type:   baseURI + "/internal",
		Title:  "Internal Server Error",
		Status: http.StatusInternalServerError,
		Detail: detail,
	}
}

// --- Fiber ErrorHandler ---

// Handler returns a Fiber ErrorHandler that serializes all errors as RFC 9457
// Problem Details. It handles three cases: *Error (our errors), *fiber.Error
// (Fiber's built-in errors like 404), and unknown errors (wrapped as 500).
func Handler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c fiber.Ctx, err error) error {
		requestID, _ := c.Locals("requestid").(string)
		path := c.Path()

		var e *Error
		var fiberErr *fiber.Error

		switch {
		case errors.As(err, &e):
			cp := *e
			cp.Instance = path
			cp.RequestID = requestID
			e = &cp

		case errors.As(err, &fiberErr):
			e = &Error{
				Type:      baseURI + "/http-error",
				Title:     http.StatusText(fiberErr.Code),
				Status:    fiberErr.Code,
				Detail:    fiberErr.Message,
				Instance:  path,
				RequestID: requestID,
			}

		default:
			e = &Error{
				Type:      baseURI + "/internal",
				Title:     "Internal Server Error",
				Status:    http.StatusInternalServerError,
				Detail:    "An unexpected error occurred",
				Instance:  path,
				RequestID: requestID,
				Cause:     err,
			}
		}

		if e.Status >= 500 {
			logger.Error("server error", "error", e)
		}
		if e.RetryAfter > 0 {
			c.Set("Retry-After", strconv.Itoa(e.RetryAfter))
		}

		c.Set(fiber.HeaderContentType, contentType)
		return c.Status(e.Status).JSON(e)
	}
}
