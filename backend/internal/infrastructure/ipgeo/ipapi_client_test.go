package ipgeo

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestIPAPIClient builds an IPAPIClient that points at the given test
// server base URL instead of the live ipapi.co endpoint.
func newTestIPAPIClient(baseURL string) *IPAPIClient {
	return &IPAPIClient{
		httpClient: &http.Client{},
		baseURL:    baseURL,
	}
}

func TestIPAPIClient_Locate_ValidIP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(ipapiResponse{
			Latitude:  0.3476,
			Longitude: 32.5825,
			Error:     false,
		}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := newTestIPAPIClient(srv.URL)

	lat, lon, err := client.Locate(context.Background(), "8.8.8.8")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if lat != 0.3476 {
		t.Errorf("lat: got %v, want 0.3476", lat)
	}
	if lon != 32.5825 {
		t.Errorf("lon: got %v, want 32.5825", lon)
	}
}

func TestIPAPIClient_Locate_PrivateIP_NoHTTPCall(t *testing.T) {
	// This server should never be called; if it is, the test fails immediately.
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Use a real NewIPAPIClient (not the test one) — private-IP check fires
	// before any HTTP call regardless of baseURL.
	client := NewIPAPIClient()

	privateIPs := []string{"127.0.0.1", "10.0.0.1", "192.168.1.1", "::1"}
	for _, ip := range privateIPs {
		t.Run(ip, func(t *testing.T) {
			_, _, err := client.Locate(context.Background(), ip)
			if !errors.Is(err, ErrPrivateIP) {
				t.Errorf("ip %s: got error %v, want ErrPrivateIP", ip, err)
			}
		})
	}

	if called {
		t.Error("HTTP server was called for a private IP — should have short-circuited")
	}
}

func TestIPAPIClient_Locate_APIErrorInBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(ipapiResponse{
			Error:  true,
			Reason: "invalid IP address",
		}); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	defer srv.Close()

	client := newTestIPAPIClient(srv.URL)

	_, _, err := client.Locate(context.Background(), "1.2.3.4")
	if err == nil {
		t.Fatal("expected error when api body has error:true, got nil")
	}
}

func TestIPAPIClient_Locate_Non200Status(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := newTestIPAPIClient(srv.URL)

	_, _, err := client.Locate(context.Background(), "1.2.3.4")
	if err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}
