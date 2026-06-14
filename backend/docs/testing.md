---
topic: testing
last_verified: 2026-06-15
sources:
  - internal/infrastructure/database/postgres/health_repository_test.go
  - internal/transport/handlers/hello_handler_test.go
  - internal/transport/handlers/health_handler_test.go
  - internal/transport/middleware/logger_test.go
  - internal/usecase/health_usecase_test.go
  - internal/infrastructure/cache/redis/cache_test.go
  - internal/bootstrap/bootstrap_test.go
---

# Testing

## Philosophy

No mocks for the database or cache. All infrastructure tests run against real services spun up by Testcontainers. This catches schema mismatches, query errors, and type coercion issues that mocks hide.

Tests MUST be independent. No test may rely on execution order or on state left by another test. Each test is responsible for creating its own fixtures.

## Test placement

| Layer | Package | Mock strategy |
|---|---|---|
| `usecase/` | `package usecase` | Local mock of repository interface |
| `transport/handlers/` | `package handlers` | Local mock of use case interface |
| `transport/middleware/` | `package middleware` | None |
| `infrastructure/database/postgres/` | `package postgres` | None — Testcontainers |
| `infrastructure/cache/redis/` | `package redis` | None — Testcontainers |
| `bootstrap/` | `package bootstrap` | Mock Pinger, `t.Setenv` |

---

## Usecase unit tests

Location: `internal/usecase/`
Package: `package usecase`
No Docker required.

Define a local struct implementing the repository interface in the same `_test.go` file. Use table-driven tests.

```go
// mockHealthReader is a local test double implementing HealthReader.
type mockHealthReader struct {
    stats domain.HealthStats
    err   error
}

func (m *mockHealthReader) Health(_ context.Context) (domain.HealthStats, error) {
    return m.stats, m.err
}

func TestGetHealth_Messages(t *testing.T) {
    tests := []struct {
        name    string
        input   domain.HealthStats
        wantMsg string
    }{
        {"default healthy message", domain.HealthStats{OpenConnections: 5}, "It's healthy"},
        {"high open connections", domain.HealthStats{OpenConnections: 41}, "The database is experiencing heavy load."},
    }

    for _, tc := range tests {
        t.Run(tc.name, func(t *testing.T) {
            uc := NewHealthUseCase(&mockHealthReader{stats: tc.input})
            got, err := uc.GetHealth(context.Background())
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if got.Message != tc.wantMsg {
                t.Errorf("message mismatch\n  got:  %q\n  want: %q", got.Message, tc.wantMsg)
            }
        })
    }
}

func TestGetHealth_RepoErrorPropagates(t *testing.T) {
    sentinel := errors.New("db unavailable")
    uc := NewHealthUseCase(&mockHealthReader{err: sentinel})
    _, err := uc.GetHealth(context.Background())
    if !errors.Is(err, sentinel) {
        t.Errorf("expected sentinel error, got: %v", err)
    }
}
```

---

## Handler unit tests

Location: `internal/transport/handlers/`
Package: `package handlers`
No Docker required.

Set `gin.SetMode(gin.TestMode)` once in an `init()` function so all tests in the package share it. Define a local struct implementing the use case interface.

```go
func init() {
    gin.SetMode(gin.TestMode)
}

// mockHealthUC is a local test double implementing usecase.HealthUseCase.
type mockHealthUC struct {
    stats domain.HealthStats
    err   error
}

func (m *mockHealthUC) GetHealth(_ context.Context) (domain.HealthStats, error) {
    return m.stats, m.err
}

func TestHealthHandler_Success(t *testing.T) {
    want := domain.HealthStats{Status: "up", Message: "It's healthy"}
    h := NewHandler(&mockHealthUC{stats: want})

    r := gin.New()
    r.GET("/health", h.HealthHandler)

    req, _ := http.NewRequest(http.MethodGet, "/health", nil)
    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", rr.Code)
    }
}

func TestHealthHandler_ServiceUnavailable(t *testing.T) {
    h := NewHandler(&mockHealthUC{err: errors.New("connection refused")})

    r := gin.New()
    r.GET("/health", h.HealthHandler)

    req, _ := http.NewRequest(http.MethodGet, "/health", nil)
    rr := httptest.NewRecorder()
    r.ServeHTTP(rr, req)

    if rr.Code != http.StatusServiceUnavailable {
        t.Errorf("expected status 503, got %d", rr.Code)
    }
}
```

---

## Redis cache integration tests

Location: `internal/infrastructure/cache/redis/`
Package: `package redis`
Requires Docker.

Uses `testcontainers.GenericContainer` with `redis:7-alpine`. The same `TestMain` + `mustStart*` pattern as PostgreSQL tests applies.

```go
var testCache usecase.CacheService

func mustStartRedisContainer() (func(context.Context, ...testcontainers.TerminateOption) error, error) {
    ctx := context.Background()

    container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image:        "redis:7-alpine",
            ExposedPorts: []string{"6379/tcp"},
            WaitingFor:   wait.ForLog("Ready to accept connections"),
        },
        Started: true,
    })
    if err != nil {
        return nil, fmt.Errorf("start redis container: %w", err)
    }

    host, _ := container.Host(ctx)
    port, _ := container.MappedPort(ctx, "6379/tcp")

    cache, err := New(fmt.Sprintf("redis://%s:%s", host, port.Port()))
    if err != nil {
        return container.Terminate, fmt.Errorf("new redis cache: %w", err)
    }
    testCache = cache

    return container.Terminate, nil
}

func TestMain(m *testing.M) {
    teardown, err := mustStartRedisContainer()
    if err != nil {
        log.Fatalf("could not start redis container: %v", err)
    }
    m.Run()
    if teardown != nil {
        if err := teardown(context.Background()); err != nil {
            log.Fatalf("could not teardown redis container: %v", err)
        }
    }
}
```

Each test operates on distinct keys so tests remain independent of each other:

```go
func TestSetAndGet(t *testing.T) {
    ctx := context.Background()
    if err := testCache.Set(ctx, "test:set-get", "hello", time.Minute); err != nil {
        t.Fatalf("Set() returned error: %v", err)
    }
    val, found, err := testCache.Get(ctx, "test:set-get")
    if err != nil {
        t.Fatalf("Get() returned error: %v", err)
    }
    if !found || val != "hello" {
        t.Errorf("expected found=true and val=%q, got found=%v val=%q", "hello", found, val)
    }
}
```

---

## Bootstrap unit tests

Location: `internal/bootstrap/`
Package: `package bootstrap`
No Docker required.

Tests access unexported functions (`loadConfig`, `validateConfig`, `jitteredBackoff`, `probeWithRetry`) directly because the test is in `package bootstrap`. Use `t.Setenv()` to set env vars — Go restores them automatically after each test. Suppress log output with `slog.NewTextHandler(io.Discard, nil)`. Mock the `Pinger` interface locally.

```go
func discardLogger() *slog.Logger {
    return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestLoadConfig_DefaultPort(t *testing.T) {
    t.Setenv("PORT", "")
    cfg := loadConfig()
    if cfg.Port != 8080 {
        t.Errorf("expected default port 8080, got %d", cfg.Port)
    }
}

func TestValidateConfig_AllMissing(t *testing.T) {
    err := validateConfig(Config{}, discardLogger())
    var cfgErr *ConfigError
    if !errors.As(err, &cfgErr) {
        t.Fatalf("expected *ConfigError, got %T: %v", err, err)
    }
    if len(cfgErr.Issues) != 5 {
        t.Errorf("expected 5 issues, got %d: %v", len(cfgErr.Issues), cfgErr.Issues)
    }
}

// mockPinger is a local test double implementing Pinger.
type mockPinger struct {
    results []error
    calls   int
}

func (m *mockPinger) PingContext(_ context.Context) error {
    if m.calls >= len(m.results) {
        return m.results[len(m.results)-1]
    }
    err := m.results[m.calls]
    m.calls++
    return err
}

func TestProbeWithRetry_SuccessAfterRetries(t *testing.T) {
    sentinel := errors.New("not ready")
    p := &mockPinger{results: []error{sentinel, sentinel, nil}}
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    if err := probeWithRetry(ctx, "test-service", p, discardLogger()); err != nil {
        t.Errorf("expected nil after eventual success, got: %v", err)
    }
}
```

---

## Postgres Testcontainers setup

`mustStartPostgresContainer()` starts a `postgres:latest` container, calls `NewPostgresDB(cfg)` with the mapped host/port, and assigns the result to the package-level `var testDB *sql.DB`.

```go
var testDB *sql.DB

func mustStartPostgresContainer() (func(context.Context, ...testcontainers.TerminateOption) error, error) {
    container, err := tcpostgres.Run(
        context.Background(),
        "postgres:latest",
        tcpostgres.WithDatabase("database"),
        tcpostgres.WithUsername("user"),
        tcpostgres.WithPassword("password"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready to accept connections").
                WithOccurrence(2).
                WithStartupTimeout(5*time.Second),
        ),
    )
    // ...resolve host and mapped port from container...
    cfg := DBConfig{Host: dbHost, Port: dbPort.Port(), ...}
    db, err := NewPostgresDB(cfg)
    testDB = db
    return container.Terminate, nil
}

func TestMain(m *testing.M) {
    teardown, err := mustStartPostgresContainer()
    if err != nil {
        log.Fatalf("could not start postgres container: %v", err)
    }
    m.Run()
    if teardown != nil && teardown(context.Background()) != nil {
        log.Fatalf("could not teardown postgres container: %v", err)
    }
}
```

## Adding a new integration test

1. Add a `TestXxx(t *testing.T)` function in a `_test.go` file under `internal/infrastructure/database/postgres/` (same package).
2. Construct the repository under test using `testDB`: e.g. `repo := NewHealthRepository(testDB)`.
3. Call repository methods directly and assert on the results.
4. Use table-driven tests for multiple cases.
5. Each test must set up its own data and not assume anything from other tests.

---

## Running tests

```bash
make test    # unit + integration (requires Docker)
make itest   # integration only — runs ./internal/infrastructure/database/postgres/...
go test ./internal/infrastructure/database/postgres/... -v -run TestHealth  # single test
```

## Requirements

- Docker must be running for any integration test (Postgres and Redis containers).
- `go test` runs all tests including integration. Use build tags if you need to separate them in the future.
