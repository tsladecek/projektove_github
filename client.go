package projektove

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"
)

type RetryConfig struct {
	MaxRetries  int
	InitialWait time.Duration
	MaxWait     time.Duration
	ShouldRetry func(statusCode int) bool
}

func (r RetryConfig) withDefaults() RetryConfig {
	if r.MaxRetries <= 0 {
		r.MaxRetries = 3
	}
	if r.InitialWait <= 0 {
		r.InitialWait = time.Second
	}
	if r.MaxWait <= 0 {
		r.MaxWait = 30 * time.Second
	}
	if r.ShouldRetry == nil {
		r.ShouldRetry = func(status int) bool { return status >= 500 }
	}
	return r
}

type Client struct {
	httpClient *http.Client
	retry      RetryConfig
}

type ClientOption func(*Client)

func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = client
	}
}

func WithRetryConfig(cfg RetryConfig) ClientOption {
	return func(c *Client) {
		c.retry = cfg
	}
}

func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, opt := range opts {
		opt(c)
	}
	c.retry = c.retry.withDefaults()
	return c
}

func (c *Client) Do(ctx context.Context, method, url string, body, target any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	if bodyReader != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.DoRaw(req, target)
}

func (c *Client) DoRaw(req *http.Request, target any) error {
	req = req.Clone(req.Context())

	var lastErr error

	for attempt := 0; attempt <= c.retry.MaxRetries; attempt++ {
		if attempt > 0 {
			if req.Body != nil && req.GetBody == nil {
				return fmt.Errorf("request body is not replayable, cannot retry")
			}
			if req.GetBody != nil {
				body, err := req.GetBody()
				if err != nil {
					return fmt.Errorf("get request body: %w", err)
				}
				req.Body = body
				req.ContentLength = -1
			}
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %w", err)
			if attempt < c.retry.MaxRetries {
				c.wait(req.Context(), attempt)
			}
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("read response body: %w", err)
			if attempt < c.retry.MaxRetries {
				c.wait(req.Context(), attempt)
			}
			continue
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if target != nil {
				if err := json.Unmarshal(respBody, target); err != nil {
					return fmt.Errorf("unmarshal response: %w", err)
				}
			}
			return nil
		}

		if c.retry.ShouldRetry(resp.StatusCode) {
			lastErr = fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
			if attempt < c.retry.MaxRetries {
				c.wait(req.Context(), attempt)
			}
			continue
		}

		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

func (c *Client) wait(ctx context.Context, attempt int) {
	wait := min(c.retry.InitialWait*(1<<attempt), c.retry.MaxWait)

	jitter := time.Duration(rand.Int63n(int64(wait/2))) - wait/4
	wait += jitter
	if wait < 0 {
		wait = 0
	}

	select {
	case <-ctx.Done():
	case <-time.After(wait):
	}
}
