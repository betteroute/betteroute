# errs

RFC 9457 Problem Details for HTTP API error responses.

## Usage

### In handlers — return errors directly

```go
func (h *Handler) Create(c fiber.Ctx) error {
    var req CreateRequest
    if err := c.Bind().JSON(&req); err != nil {
        return errs.BadRequest("invalid json")
    }
    if req.URL == "" {
        return errs.BadRequest("url is required")
    }

    link, err := h.service.Create(c.Context(), req)
    if err != nil {
        return h.mapError(err)
    }
    return c.Status(fiber.StatusCreated).JSON(link)
}
```

### In services — return sentinel errors

```go
// errors.go
var (
    ErrLinkNotFound = errors.New("link not found")
    ErrSlugTaken    = errors.New("slug already taken")
)

// service.go
func (s *Service) Create(ctx context.Context, input CreateInput) (*Link, error) {
    if _, err := s.store.GetBySlug(ctx, input.Slug); err == nil {
        return nil, ErrSlugTaken
    }
    return s.store.Create(ctx, input)
}
```

### In stores — translate DB errors to sentinel errors

```go
// store.go
func (s *Store) GetBySlug(ctx context.Context, slug string) (*Link, error) {
    row, err := s.q.GetLinkBySlug(ctx, slug)
    if errors.Is(err, pgx.ErrNoRows) {
        return nil, ErrLinkNotFound
    }
    if err != nil {
        return nil, err
    }
    return &row, nil
}
```

### In handlers — map domain errors to HTTP errors

```go
func (h *Handler) mapError(err error) error {
    switch {
    case errors.Is(err, ErrLinkNotFound):
        return errs.NotFound("Link", "")
    case errors.Is(err, ErrSlugTaken):
        return errs.Conflict("slug already taken")
    default:
        return errs.Internal("").WithCause(err)
    }
}
```

### Attach internal cause (logged, never sent to client)

```go
return errs.Internal("").WithCause(err)
// Client sees: {"detail": "An unexpected error occurred"}
// Server logs: {"cause": "pq: connection refused"}
```

### Validation errors

```go
return errs.Validation([]errs.FieldError{
    {Field: "email", Message: "must be a valid email"},
    {Field: "name", Message: "must be 1-100 characters"},
})
```

## Available Constructors

| Function | Status | When to use |
|----------|--------|-------------|
| `BadRequest(detail)` | 400 | Malformed request, missing fields |
| `Unauthorized(detail)` | 401 | Invalid/missing auth credentials |
| `Forbidden(detail)` | 403 | Valid auth but insufficient permissions |
| `NotFound(resource, id)` | 404 | Resource doesn't exist |
| `Conflict(detail)` | 409 | Duplicate resource, already exists |
| `Validation(fields)` | 422 | Field-level validation failures |
| `TooManyRequests(retryAfter)` | 429 | Rate limit exceeded |
| `Internal(detail)` | 500 | Unexpected server error |

## Response Format

All errors are serialized as `application/problem+json` per [RFC 9457](https://www.rfc-editor.org/rfc/rfc9457):

```json
{
  "type": "https://api.betteroute.dev/errors/not-found",
  "title": "Not Found",
  "status": 404,
  "detail": "Link 'abc123' not found",
  "instance": "/api/links/abc123",
  "request_id": "req_xyz"
}
```

## Architecture

```
errors.go   → sentinel errors (domain-level, no HTTP knowledge)
store.go    → translates DB errors (pgx.ErrNoRows) → sentinel errors
service.go  → returns sentinel errors
handler.go  → mapError() converts domain errors → errs.*() HTTP errors
errs pkg    → ErrorHandler serializes as RFC 9457 JSON
```

Stores and services **never** import `errs`. Only handlers translate between layers.
