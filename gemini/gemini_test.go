package gemini

import (
	"context"
	"testing"
	"time"
)

func TestExponentialBackoff(t *testing.T) {
	client := NewClient("fake-key")

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 500 * time.Millisecond},
		{1, 1 * time.Second},
		{2, 2 * time.Second},
		{3, 4 * time.Second},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := client.exponentialBackoff(tt.attempt)
			if result != tt.expected {
				t.Errorf("exponentialBackoff(%d) = %v, want %v", tt.attempt, result, tt.expected)
			}
		})
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "network error",
			err:  testError("network error: connection refused"),
			want: true,
		},
		{
			name: "429 rate limit",
			err:  testError("gemini error (429): quota exceeded"),
			want: true,
		},
		{
			name: "500 server error",
			err:  testError("gemini error (500): internal server error"),
			want: true,
		},
		{
			name: "503 service unavailable",
			err:  testError("gemini error (503): service unavailable"),
			want: true,
		},
		{
			name: "400 bad request",
			err:  testError("gemini error (400): invalid request"),
			want: false,
		},
		{
			name: "401 unauthorized",
			err:  testError("gemini error (401): invalid API key"),
			want: false,
		},
		{
			name: "403 forbidden",
			err:  testError("gemini error (403): access denied"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryable(tt.err)
			if result != tt.want {
				t.Errorf("isRetryable(%v) = %v, want %v", tt.err, result, tt.want)
			}
		})
	}
}

func TestNewClientWithLimits(t *testing.T) {
	tests := []struct {
		name              string
		requestsPerSecond int
		retryDelay        time.Duration
		wantMaxRetries    int
		wantRetryDelay    time.Duration
	}{
		{
			name:              "default 100 req/sec",
			requestsPerSecond: 100,
			retryDelay:        500 * time.Millisecond,
			wantMaxRetries:    3,
			wantRetryDelay:    500 * time.Millisecond,
		},
		{
			name:              "50 req/sec",
			requestsPerSecond: 50,
			retryDelay:        1 * time.Second,
			wantMaxRetries:    3,
			wantRetryDelay:    1 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClientWithLimits("fake-key", tt.requestsPerSecond, tt.retryDelay)

			if client.maxRetries != tt.wantMaxRetries {
				t.Errorf("maxRetries = %d, want %d", client.maxRetries, tt.wantMaxRetries)
			}
			if client.retryDelay != tt.wantRetryDelay {
				t.Errorf("retryDelay = %v, want %v", client.retryDelay, tt.wantRetryDelay)
			}
		})
	}
}

func TestContextCancellation(t *testing.T) {
	client := NewClient("fake-key")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Generate(ctx, "test prompt")
	if err != context.Canceled {
		t.Errorf("Generate with canceled context = %v, want context.Canceled", err)
	}
}

type testError string

func (e testError) Error() string {
	return string(e)
}
