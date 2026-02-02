<!-- Parent: ../AGENTS.md -->
<!-- Generated: 2026-02-03 | Updated: 2026-02-03 -->

# middleware

## Purpose

HTTP middleware for cross-cutting concerns: logging, recovery, rate limiting, future authentication.

## Key Files

| File | Description |
|------|-------------|
| `logging.go` | Request/response logging |
| `recovery.go` | Panic recovery |
| `ratelimit.go` | Rate limiting (optional) |
| `auth.go` | Authentication (future) |

## For AI Agents

### Working In This Directory

- Middleware follows `func(http.Handler) http.Handler` pattern
- Chain middleware in order: Recovery -> Logging -> Auth -> Handler
- Use `context.WithValue` sparingly for request-scoped data
- Log structured data (JSON) for production

### Middleware Pattern

```go
func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        // Wrap response writer to capture status
        ww := &responseWriter{ResponseWriter: w, status: 200}

        next.ServeHTTP(ww, r)

        log.Printf("method=%s path=%s status=%d duration=%s",
            r.Method, r.URL.Path, ww.status, time.Since(start))
    })
}

type responseWriter struct {
    http.ResponseWriter
    status int
}

func (w *responseWriter) WriteHeader(status int) {
    w.status = status
    w.ResponseWriter.WriteHeader(status)
}
```

### Recovery Middleware

```go
func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic recovered: %v\n%s", err, debug.Stack())
                http.Error(w, "Internal Server Error", http.StatusInternalServerError)
            }
        }()
        next.ServeHTTP(w, r)
    })
}
```

### Future: Authentication

```go
// Token-based auth (v2)
func Auth(tokenValidator TokenValidator) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := r.Header.Get("Authorization")
            if !tokenValidator.Valid(token) {
                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

<!-- MANUAL: -->
