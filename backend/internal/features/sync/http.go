package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// HTTPClient talks to the self-hosted sync server.
type HTTPClient struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

// NewHTTPClient returns a client for a sync server. baseURL must not
// carry a trailing slash.
func NewHTTPClient(baseURL, apiKey string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    &http.Client{Timeout: 60 * time.Second},
	}
}

// Register creates a device on the server and returns its credentials.
func (c *HTTPClient) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {
	var out RegisterResponse
	if err := c.do(ctx, http.MethodPost, pathRegister, "", req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// Sync performs a bidirectional replication exchange.
func (c *HTTPClient) Sync(ctx context.Context, deviceToken string, req *Request) (*Response, error) {
	var out Response
	if err := c.do(ctx, http.MethodPost, pathSync, deviceToken, req, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *HTTPClient) do(ctx context.Context, method, path, deviceToken string, body, out any) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return fmt.Errorf("sync client: encode request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, &buf)
	if err != nil {
		return fmt.Errorf("sync client: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", c.apiKey)
	if deviceToken != "" {
		httpReq.Header.Set("X-Device-Token", deviceToken)
	}
	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("sync client: %s %s: %w", method, path, err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sync client: %s %s: status %d: %s", method, path, resp.StatusCode, trimBody(raw))
	}
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			return fmt.Errorf("sync client: decode %s %s: %w", method, path, err)
		}
	}
	return nil
}

func trimBody(b []byte) string {
	if len(b) > 500 {
		b = b[:500]
	}
	return string(b)
}
