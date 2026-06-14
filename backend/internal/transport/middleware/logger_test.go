package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
	// Discard all slog output so test logs stay clean.
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))
}

// newRouter returns a gin.Engine with the Logger middleware attached and a
// single GET route at path that calls the provided handler.
func newRouter(path string, handler gin.HandlerFunc) *gin.Engine {
	r := gin.New()
	r.Use(Logger())
	r.GET(path, handler)
	return r
}

func TestLogger_PassesThroughResponse(t *testing.T) {
	tests := []struct {
		name           string
		handlerStatus  int
		wantStatusCode int
	}{
		{"2xx passes through", http.StatusOK, http.StatusOK},
		{"4xx passes through", http.StatusNotFound, http.StatusNotFound},
		{"5xx passes through", http.StatusInternalServerError, http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newRouter("/test", func(c *gin.Context) {
				c.Status(tc.handlerStatus)
			})

			req, err := http.NewRequest(http.MethodGet, "/test", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			if rr.Code != tc.wantStatusCode {
				t.Errorf("expected status %d, got %d", tc.wantStatusCode, rr.Code)
			}
		})
	}
}

func TestLogger_WithQueryString(t *testing.T) {
	r := newRouter("/search", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req, err := http.NewRequest(http.MethodGet, "/search?q=hello&limit=10", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// Must not panic.
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rr.Code)
	}
}

func TestLogger_WithGinErrors(t *testing.T) {
	r := newRouter("/err", func(c *gin.Context) {
		// Attach a gin error so the middleware's error-logging branch is exercised.
		_ = c.Error(http.ErrBodyNotAllowed)
		c.Status(http.StatusBadRequest)
	})

	req, err := http.NewRequest(http.MethodGet, "/err", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()

	// Must not panic even with c.Errors populated.
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", rr.Code)
	}
}

func TestLogger_DoesNotModifyBody(t *testing.T) {
	r := newRouter("/body", func(c *gin.Context) {
		c.String(http.StatusOK, "hello")
	})

	req, err := http.NewRequest(http.MethodGet, "/body", nil)
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if body := rr.Body.String(); body != "hello" {
		t.Errorf("expected body %q, got %q", "hello", body)
	}
}
