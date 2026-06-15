package ipgeo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// ErrPrivateIP is returned when the caller passes a loopback or RFC-1918
// address. Callers should silently skip the location update in this case.
var ErrPrivateIP = errors.New("ipgeo: private or loopback IP address")

// IPAPIClient implements domain/services.IPGeolocationService by querying
// https://ipapi.co/{ip}/json/.
type IPAPIClient struct {
	httpClient *http.Client
	// baseURL is overridable in tests; production code leaves it empty and the
	// default "https://ipapi.co" is used.
	baseURL string
}

// NewIPAPIClient returns an IPAPIClient with a 5-second HTTP client timeout.
func NewIPAPIClient() *IPAPIClient {
	return &IPAPIClient{httpClient: &http.Client{Timeout: 5 * time.Second}}
}

// ipapiResponse matches the JSON shape returned by ipapi.co.
type ipapiResponse struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Error     bool    `json:"error"`
	Reason    string  `json:"reason"`
}

// Locate resolves ip to approximate lat/lon coordinates.
// Returns ErrPrivateIP for loopback and RFC-1918 addresses without making
// an outbound HTTP call.
func (c *IPAPIClient) Locate(ctx context.Context, ip string) (lat, lon float64, err error) {
	if isPrivateIP(ip) {
		return 0, 0, ErrPrivateIP
	}

	base := c.baseURL
	if base == "" {
		base = "https://ipapi.co"
	}
	url := fmt.Sprintf("%s/%s/json/", base, ip)
	return c.fetch(ctx, url)
}

// fetch performs the HTTP GET and parses the response.
func (c *IPAPIClient) fetch(ctx context.Context, url string) (lat, lon float64, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, 0, fmt.Errorf("ipgeo: build request: %w", err)
	}
	req.Header.Set("User-Agent", "gigz-backend/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, 0, fmt.Errorf("ipgeo: http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("ipgeo: unexpected status %d", resp.StatusCode)
	}

	var body ipapiResponse
	if decErr := json.NewDecoder(resp.Body).Decode(&body); decErr != nil {
		return 0, 0, fmt.Errorf("ipgeo: decode response: %w", decErr)
	}

	if body.Error {
		return 0, 0, fmt.Errorf("ipgeo: api error: %s", body.Reason)
	}

	return body.Latitude, body.Longitude, nil
}

// isPrivateIP reports whether the given IP string is a loopback or
// RFC-1918 / RFC-4193 private address.
func isPrivateIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate()
}
