package redis

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"backend/internal/usecase"
)

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

	host, err := container.Host(ctx)
	if err != nil {
		return container.Terminate, fmt.Errorf("container host: %w", err)
	}

	port, err := container.MappedPort(ctx, "6379/tcp")
	if err != nil {
		return container.Terminate, fmt.Errorf("mapped port: %w", err)
	}

	redisURL := fmt.Sprintf("redis://%s:%s", host, port.Port())

	cache, err := New(redisURL)
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

func TestPingContext(t *testing.T) {
	err := testCache.PingContext(context.Background())
	if err != nil {
		t.Errorf("PingContext() returned unexpected error: %v", err)
	}
}

func TestGet_MissingKey(t *testing.T) {
	ctx := context.Background()
	val, found, err := testCache.Get(ctx, "does-not-exist")
	if err != nil {
		t.Fatalf("Get() on missing key returned error: %v", err)
	}
	if found {
		t.Error("Get() on missing key: expected found=false, got true")
	}
	if val != "" {
		t.Errorf("Get() on missing key: expected empty string, got %q", val)
	}
}

func TestSetAndGet(t *testing.T) {
	ctx := context.Background()
	key := "test:set-get"

	if err := testCache.Set(ctx, key, "hello", time.Minute); err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}

	val, found, err := testCache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if !found {
		t.Fatal("Get() after Set(): expected found=true, got false")
	}
	if val != "hello" {
		t.Errorf("Get() after Set(): expected %q, got %q", "hello", val)
	}
}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	key := "test:delete"

	if err := testCache.Set(ctx, key, "to-delete", time.Minute); err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}

	if err := testCache.Delete(ctx, key); err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}

	_, found, err := testCache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() after Delete() returned error: %v", err)
	}
	if found {
		t.Error("Get() after Delete(): expected found=false, got true")
	}
}

func TestExists(t *testing.T) {
	ctx := context.Background()
	key := "test:exists"

	exists, err := testCache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists() on missing key returned error: %v", err)
	}
	if exists {
		t.Error("Exists() on missing key: expected false, got true")
	}

	if err := testCache.Set(ctx, key, "present", time.Minute); err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}

	exists, err = testCache.Exists(ctx, key)
	if err != nil {
		t.Fatalf("Exists() after Set() returned error: %v", err)
	}
	if !exists {
		t.Error("Exists() after Set(): expected true, got false")
	}
}

func TestSetNX(t *testing.T) {
	ctx := context.Background()
	key := "test:setnx"

	// First call on a missing key must succeed.
	ok, err := testCache.SetNX(ctx, key, "first", time.Minute)
	if err != nil {
		t.Fatalf("SetNX() first call returned error: %v", err)
	}
	if !ok {
		t.Error("SetNX() first call: expected true (key was absent), got false")
	}

	// Second call on the same key must fail (key already exists).
	ok, err = testCache.SetNX(ctx, key, "second", time.Minute)
	if err != nil {
		t.Fatalf("SetNX() second call returned error: %v", err)
	}
	if ok {
		t.Error("SetNX() second call: expected false (key already present), got true")
	}

	// The stored value must still be the one from the first call.
	val, found, err := testCache.Get(ctx, key)
	if err != nil {
		t.Fatalf("Get() after SetNX() returned error: %v", err)
	}
	if !found {
		t.Fatal("Get() after SetNX(): expected found=true")
	}
	if val != "first" {
		t.Errorf("Get() after SetNX(): expected %q, got %q", "first", val)
	}
}

func TestClose(t *testing.T) {
	// Create a separate cache instance so closing it doesn't affect other tests.
	// The container is still running; we just close the client connection.
	ctx := context.Background()

	// Obtain the host/port from the existing test container by pinging — if
	// testCache is healthy its address is reachable. We create a fresh client.
	// Rather than re-introspect the container we rely on the exported New()
	// constructor with the same URL. Since testCache is still open and the
	// container is up, we can derive the address by parsing the internal client.
	// The simpler approach: call Close on a freshly-constructed instance.
	//
	// We need the URL. Because New() accepts a URL string we capture it indirectly:
	// ping testCache to confirm it's still healthy, then create a second client
	// that we can safely close.
	if err := testCache.PingContext(ctx); err != nil {
		t.Fatalf("testCache is not healthy before Close test: %v", err)
	}

	// Reconstruct the URL from the container — we do this by casting to the
	// concrete type to reach the underlying redis.Client options.
	cs, ok := testCache.(*cacheService)
	if !ok {
		t.Fatal("testCache is not *cacheService")
	}
	addr := cs.client.Options().Addr
	secondCache, err := New("redis://" + addr)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if err := secondCache.Close(); err != nil {
		t.Errorf("Close() returned unexpected error: %v", err)
	}
}
