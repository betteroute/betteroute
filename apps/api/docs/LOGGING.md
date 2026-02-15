# Logging

We use Go's built-in `log/slog` — no wrapper package, no third-party library.

## Setup

Logger is configured in `cmd/api/main.go`:

- **Development**: `TextHandler`, debug level, source locations enabled
- **Production**: `JSONHandler`, info level

## Rules

### Use structured key-value pairs, not string interpolation

```go
// ✅ Good
logger.Info("link created", "id", link.ID, "slug", link.Slug)

// ❌ Bad
logger.Info(fmt.Sprintf("link %s created with slug %s", link.ID, link.Slug))
```

### Use static message strings

```go
// ✅ Good — sloglint enforces this
logger.Error("failed to create link", "error", err)

// ❌ Bad — dynamic message
logger.Error("failed to create link: " + err.Error())
```

### Pass logger via struct injection, not context

```go
// ✅ Logger is a struct dependency, not pulled from context.
type Handler struct {
    logger  *slog.Logger
    service *Service
}
```

### Log levels

| Level | When |
|-------|------|
| `Debug` | Request/response details, cache hits/misses, query timing |
| `Info` | Server start/stop, route registration, successful operations |
| `Warn` | Recoverable issues, degraded functionality, retries |
| `Error` | Failed operations, unhandled errors (auto-logged by `errs.Handler` for 5xx) |

### Don't log what the error handler already logs

5xx errors are automatically logged by `errs.Handler(logger)` with status, detail, and cause.
No need to log them again in your handler.

```go
// ✅ Good — just return, ErrorHandler logs it
return errs.Internal("").WithCause(err)

// ❌ Bad — double logging
logger.Error("something failed", "error", err)
return errs.Internal("").WithCause(err)
```

## Future additions

- **Redaction**: `ReplaceAttr` for sensitive keys (password, token, etc.) — add when auth is wired
- **Request ID enrichment**: Add request ID to all logs via middleware — add when request ID middleware exists
