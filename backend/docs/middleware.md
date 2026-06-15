---
topic: middleware
last_verified: 2026-06-15
sources:
  - internal/transport/middleware/logger.go
  - internal/transport/middleware/ratelimit.go
  - internal/transport/handlers/routes.go
---

# Middleware

All middleware lives in `internal/transport/middleware/` and follows the Gin `HandlerFunc` pattern. Middleware is registered in `RegisterRoutes()` inside `internal/transport/handlers/routes.go`.

## Registration order

```go
// 1. Recovery + logger (debug: gin.Logger, release: middleware.Logger)
r.Use(gin.Recovery(), middleware.Logger())
// 2. Rate limiter (no-op when RPS <= 0)
r.Use(middleware.RateLimit(rps, burst))
// 3. CORS
r.Use(cors.New(...))
```

## Logger

`Logger() gin.HandlerFunc` emits one structured `slog` record per request after `c.Next()` returns. Fields: `status`, `method`, `path`, `latency`, `ip`, and optionally `query` and `errors`.

In debug mode (`ENV` not set to `staging`/`production`) Gin's built-in colorful logger is used instead.

## Rate limiter

`RateLimit(rps float64, burst int) gin.HandlerFunc` limits each client IP to `rps` requests per second using a token-bucket algorithm (`golang.org/x/time/rate`). Each IP gets its own `rate.Limiter` stored in a mutex-guarded map.

- **Disabled** when `rps <= 0` (no-op middleware returned).
- **429 Too Many Requests** returned when the bucket is empty; body: `{"error": "rate limit exceeded"}`.
- Configured via env vars `RATE_LIMIT_RPS` and `RATE_LIMIT_BURST` (see [environment](environment.md)).

## Adding new middleware

1. Create `internal/transport/middleware/<name>.go` with a function returning `gin.HandlerFunc`.
2. Register it in `RegisterRoutes()` in `internal/transport/handlers/routes.go` at the appropriate position in the chain.
3. If it requires configuration, add fields to `bootstrap.Config` and read from env in `loadConfig()`.
