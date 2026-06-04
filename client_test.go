package projektove

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testPayload struct {
	Key string `json:"key"`
}

func TestClient_Do(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		setup func(t *testing.T) (*Client, *http.Request, any)
		check func(t *testing.T, err error)
	}{
		{
			name: "GET success",
			setup: func(t *testing.T) (*Client, *http.Request, any) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "GET", r.Method)
					json.NewEncoder(w).Encode(testPayload{Key: "value"})
				}))
				t.Cleanup(ts.Close)

				req, _ := http.NewRequestWithContext(context.Background(), "GET", ts.URL, nil)
				var target testPayload
				return NewClient(), req, &target
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "POST with body",
			setup: func(t *testing.T) (*Client, *http.Request, any) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "POST", r.Method)
					var got testPayload
					json.NewDecoder(r.Body).Decode(&got)
					json.NewEncoder(w).Encode(got)
				}))
				t.Cleanup(ts.Close)

				body, _ := json.Marshal(testPayload{Key: "posted"})
				req, _ := http.NewRequestWithContext(context.Background(), "POST", ts.URL, bytes.NewReader(body))
				var target testPayload
				return NewClient(), req, &target
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "4xx no retry",
			setup: func(t *testing.T) (*Client, *http.Request, any) {
				var reqCount atomic.Int32
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					reqCount.Add(1)
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`bad request`))
				}))
				t.Cleanup(ts.Close)

				req, _ := http.NewRequestWithContext(context.Background(), "GET", ts.URL, nil)
				client := NewClient()
				return client, req, new(struct{})
			},
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "400")
				assert.Contains(t, err.Error(), "bad request")
			},
		},
		{
			name: "5xx then success",
			setup: func(t *testing.T) (*Client, *http.Request, any) {
				var reqCount atomic.Int32
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					c := reqCount.Add(1)
					if c == 1 {
						w.WriteHeader(http.StatusInternalServerError)
						w.Write([]byte(`server error`))
						return
					}
					json.NewEncoder(w).Encode(testPayload{Key: "ok"})
				}))
				t.Cleanup(ts.Close)

				req, _ := http.NewRequestWithContext(context.Background(), "GET", ts.URL, nil)
				client := NewClient(WithRetryConfig(RetryConfig{
					MaxRetries:  2,
					InitialWait: time.Millisecond,
				}))
				var target testPayload
				return client, req, &target
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "all retries fail",
			setup: func(t *testing.T) (*Client, *http.Request, any) {
				var reqCount atomic.Int32
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					reqCount.Add(1)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`still failing`))
				}))
				t.Cleanup(ts.Close)

				req, _ := http.NewRequestWithContext(context.Background(), "GET", ts.URL, nil)
				client := NewClient(WithRetryConfig(RetryConfig{
					MaxRetries:  2,
					InitialWait: time.Millisecond,
				}))
				return client, req, new(struct{})
			},
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "max retries exceeded")
			},
		},
		{
			name: "network error retry",
			setup: func(t *testing.T) (*Client, *http.Request, any) {
				var reqCount atomic.Int32
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					c := reqCount.Add(1)
					if c == 1 {
						hj, ok := w.(http.Hijacker)
						if !ok {
							w.WriteHeader(http.StatusInternalServerError)
							return
						}
						conn, _, err := hj.Hijack()
						require.NoError(t, err)
						conn.Close()
						return
					}
					json.NewEncoder(w).Encode(testPayload{Key: "recovered"})
				}))
				t.Cleanup(ts.Close)

				req, _ := http.NewRequestWithContext(context.Background(), "GET", ts.URL, nil)
				client := NewClient(WithRetryConfig(RetryConfig{
					MaxRetries:  2,
					InitialWait: time.Millisecond,
				}))
				var target testPayload
				return client, req, &target
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "nil target",
			setup: func(t *testing.T) (*Client, *http.Request, any) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"key":"value"}`))
				}))
				t.Cleanup(ts.Close)

				req, _ := http.NewRequestWithContext(context.Background(), "GET", ts.URL, nil)
				return NewClient(), req, nil
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "non-replayable body",
			setup: func(t *testing.T) (*Client, *http.Request, any) {
				var reqCount atomic.Int32
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					reqCount.Add(1)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`fail`))
				}))
				t.Cleanup(ts.Close)

				req, _ := http.NewRequestWithContext(context.Background(), "POST", ts.URL, nil)
				req.Body = io.NopCloser(bytes.NewReader([]byte(`{"key":"val"}`)))
				req.ContentLength = 12

				client := NewClient(WithRetryConfig(RetryConfig{
					MaxRetries:  1,
					InitialWait: time.Millisecond,
				}))
				return client, req, new(struct{})
			},
			check: func(t *testing.T, err error) {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "not replayable")
			},
		},
		{
			name: "context cancelled",
			setup: func(t *testing.T) (*Client, *http.Request, any) {
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}))
				t.Cleanup(ts.Close)

				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				req, _ := http.NewRequestWithContext(ctx, "GET", ts.URL, nil)
				client := NewClient(WithRetryConfig(RetryConfig{
					MaxRetries:  5,
					InitialWait: time.Millisecond,
				}))
				return client, req, new(struct{})
			},
			check: func(t *testing.T, err error) {
				require.Error(t, err)
			},
		},
		{
			name: "custom ShouldRetry for 429",
			setup: func(t *testing.T) (*Client, *http.Request, any) {
				var reqCount atomic.Int32
				ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					c := reqCount.Add(1)
					if c == 1 {
						w.WriteHeader(http.StatusTooManyRequests)
						w.Write([]byte(`rate limited`))
						return
					}
					json.NewEncoder(w).Encode(testPayload{Key: "ok"})
				}))
				t.Cleanup(ts.Close)

				req, _ := http.NewRequestWithContext(context.Background(), "GET", ts.URL, nil)
				client := NewClient(WithRetryConfig(RetryConfig{
					MaxRetries:  2,
					InitialWait: time.Millisecond,
					ShouldRetry: func(status int) bool { return status >= 429 },
				}))
				var target testPayload
				return client, req, &target
			},
			check: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, req, target := tt.setup(t)
			_, err := client.Do(req, target)
			tt.check(t, err)
		})
	}
}

func TestClient_Do_response_decode(t *testing.T) {
	t.Parallel()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"result": "hello"})
	}))
	t.Cleanup(ts.Close)

	req, _ := http.NewRequestWithContext(context.Background(), "GET", ts.URL, nil)

	var got struct {
		Result string `json:"result"`
	}
	_, err := NewClient().Do(req, &got)
	require.NoError(t, err)
	assert.Equal(t, "hello", got.Result)
}

func TestNew_defaults(t *testing.T) {
	c := NewClient()
	assert.NotNil(t, c.httpClient)
	assert.Equal(t, 3, c.retry.MaxRetries)
	assert.Equal(t, time.Second, c.retry.InitialWait)
	assert.Equal(t, 30*time.Second, c.retry.MaxWait)
	assert.NotNil(t, c.retry.ShouldRetry)
}

func TestRetryConfig_withDefaults(t *testing.T) {
	r := RetryConfig{}.withDefaults()
	assert.Equal(t, 3, r.MaxRetries)
	assert.Equal(t, time.Second, r.InitialWait)
	assert.Equal(t, 30*time.Second, r.MaxWait)
	assert.True(t, r.ShouldRetry(500))
	assert.True(t, r.ShouldRetry(503))
	assert.False(t, r.ShouldRetry(400))
	assert.False(t, r.ShouldRetry(429))
}

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{Timeout: 5 * time.Second}
	c := NewClient(WithHTTPClient(custom))
	assert.Same(t, custom, c.httpClient)
}

func TestWithRetryConfig(t *testing.T) {
	cfg := RetryConfig{MaxRetries: 7, InitialWait: 123 * time.Millisecond, MaxWait: time.Second}
	c := NewClient(WithRetryConfig(cfg))
	assert.Equal(t, 7, c.retry.MaxRetries)
	assert.Equal(t, 123*time.Millisecond, c.retry.InitialWait)
	assert.Equal(t, time.Second, c.retry.MaxWait)
}

func TestClient_Do_4xx_no_retry_after_partial_body_read(t *testing.T) {
	t.Parallel()

	data := []byte(`{"key":"val"}`)
	req, _ := http.NewRequestWithContext(
		context.Background(),
		"POST",
		"http://127.0.0.1:1/nonexistent",
		bytes.NewReader(data),
	)

	client := NewClient(WithRetryConfig(RetryConfig{
		MaxRetries:  2,
		InitialWait: time.Millisecond,
	}))

	_, err := client.Do(req, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max retries exceeded")
}
