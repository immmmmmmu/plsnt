package api

import (
	"context"
)

// Client abstracts communication with the Pleasanter API.
// All Pleasanter API endpoints use HTTP POST.
type Client interface {
	// Post sends a typed request and unmarshals the response into resp.
	Post(ctx context.Context, endpoint string, req any, resp any) error
	// PostRaw sends a map payload and returns a map response (for --json bypass).
	PostRaw(ctx context.Context, endpoint string, payload map[string]any) (map[string]any, error)
	// NoRetry returns a copy of this client with retry disabled.
	// Use for non-idempotent operations (create, delete).
	NoRetry() Client
}

// Option configures the client.
type Option func(*clientImpl)

// WithRetryDisabled disables retry for non-idempotent operations.
func WithRetryDisabled() Option {
	return func(c *clientImpl) {
		c.retryDisabled = true
	}
}

// WithInsecure skips TLS certificate verification.
func WithInsecure() Option {
	return func(c *clientImpl) {
		c.insecure = true
	}
}
