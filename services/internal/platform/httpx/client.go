package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// Client is a small JSON HTTP client for inter-service calls. It centralises request building,
// timeouts and the decoding of the shared error envelope so callers only handle status + result.
type Client struct {
	base string
	hc   *http.Client
}

func NewClient(base string) *Client {
	return &Client{base: base, hc: &http.Client{Timeout: 5 * time.Second}}
}

// Do sends method+path with an optional JSON body. On a 2xx it decodes into out (if non-nil) and
// returns the status with a nil error. On a non-2xx it decodes the error envelope and returns it as
// the error, so callers can inspect status (e.g. 409) and the envelope's SKU. Transport failures
// return a zero status and the underlying error.
func (c *Client) Do(ctx context.Context, method, path string, body, out any) (int, error) {
	return c.DoWithHeaders(ctx, method, path, nil, body, out)
}

// DoWithHeaders is Do plus extra request headers, used to forward the tenant (X-Tenant-ID) on
// inter-service calls so the downstream stays scoped to the same store.
func (c *Client) DoWithHeaders(ctx context.Context, method, path string, headers map[string]string, body, out any) (int, error) {
	var reader io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, err
		}
		reader = bytes.NewReader(raw)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.base+path, reader)
	if err != nil {
		return 0, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	for k, v := range headers {
		if v != "" {
			req.Header.Set(k, v)
		}
	}

	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		if out != nil {
			if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
				return resp.StatusCode, err
			}
		}
		return resp.StatusCode, nil
	}

	var env Error
	if err := json.NewDecoder(resp.Body).Decode(&env); err != nil || env.Message == "" {
		env = Error{Message: "upstream error", Detail: http.StatusText(resp.StatusCode)}
	}
	return resp.StatusCode, &env
}
