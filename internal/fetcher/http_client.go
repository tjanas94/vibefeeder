package fetcher

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/tjanas94/vibefeeder/internal/shared/ssrf"
)

// HTTP header names
const (
	HeaderUserAgent       = "User-Agent"
	HeaderIfNoneMatch     = "If-None-Match"
	HeaderIfModifiedSince = "If-Modified-Since"
	HeaderETag            = "ETag"
	HeaderLastModified    = "Last-Modified"
	HeaderCacheControl    = "Cache-Control"
	HeaderRetryAfter      = "Retry-After"
	UserAgentValue        = "VibeFeeder/1.0 (+https://github.com/tjanas94/vibefeeder; mailto:vibefeeder@janas.dev)"
)

// HTTPClientConfig contains configuration for creating an HTTP client
type HTTPClientConfig struct {
	Timeout         time.Duration
	FollowRedirects bool
	Logger          *slog.Logger
}

// HTTPClient wraps http.Client with SSRF protection and User-Agent
type HTTPClient struct {
	client *http.Client
	logger *slog.Logger
}

// Ensure HTTPClient implements HTTPClientInterface at compile time
var _ HTTPClientInterface = (*HTTPClient)(nil)

// NewHTTPClient creates a new HTTP client with SSRF protection via custom Dialer.
// The SSRF validation happens at the connection level, preventing TOCTOU vulnerabilities.
func NewHTTPClient(cfg HTTPClientConfig) *HTTPClient {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	baseDialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}

	// Create custom transport with SSRF-protected dialer
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			// Extract host from addr (format: "host:port")
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, fmt.Errorf("invalid address: %w", err)
			}

			// Resolve DNS to get IP addresses
			ips, err := net.DefaultResolver.LookupIP(ctx, "ip", host)
			if err != nil {
				return nil, fmt.Errorf("DNS resolution failed for %s: %w", host, err)
			}

			if len(ips) == 0 {
				return nil, fmt.Errorf("no IP addresses found for %s", host)
			}

			// Validate all resolved IPs against SSRF attacks
			for _, ip := range ips {
				if err := ssrf.ValidateIP(ip); err != nil {
					cfg.Logger.Error("SSRF validation failed", "host", host, "ip", ip.String(), "error", err)
					return nil, fmt.Errorf("security validation failed for %s: %w", host, err)
				}
			}

			// Use the first valid IP to dial
			validAddr := net.JoinHostPort(ips[0].String(), port)
			return baseDialer.DialContext(ctx, network, validAddr)
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	client := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}

	if !cfg.FollowRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return &HTTPClient{
		client: client,
		logger: cfg.Logger,
	}
}

// ExecuteRequestParams contains parameters for executing an HTTP request
type ExecuteRequestParams struct {
	URL          string
	ETag         *string
	LastModified *string
}

// ExecuteRequest creates and executes an HTTP GET request with optional conditional headers.
// The caller is responsible for closing the Response.Body when the response is no longer needed.
//
// The context timeout should be set appropriately for the expected request duration.
// The HTTPClient's Timeout serves as a fallback if no context timeout is set.
func (c *HTTPClient) ExecuteRequest(ctx context.Context, params ExecuteRequestParams) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", params.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set User-Agent header
	req.Header.Set(HeaderUserAgent, UserAgentValue)

	// Add conditional request headers if available
	if params.ETag != nil && *params.ETag != "" {
		req.Header.Set(HeaderIfNoneMatch, *params.ETag)
	}
	if params.LastModified != nil && *params.LastModified != "" {
		req.Header.Set(HeaderIfModifiedSince, *params.LastModified)
	}

	return c.client.Do(req)
}
