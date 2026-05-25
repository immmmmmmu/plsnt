package api

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"

	"github.com/immmmmmmu/plsnt/internal/errs"
)

type clientImpl struct {
	baseURL       string
	apiKey        string
	apiVersion    string
	httpClient    *http.Client
	retryDisabled bool
	insecure      bool
}

// New creates a new API client.
func New(baseURL, apiKey, apiVersion string, opts ...Option) Client {
	transport := &http.Transport{
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   10,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:  10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 30 * time.Second
	retryClient.Logger = nil // suppress default log output
	retryClient.HTTPClient = &http.Client{
		Transport: transport,
		Timeout:   60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // do not follow redirects for API calls
		},
	}

	c := &clientImpl{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		apiVersion: apiVersion,
		httpClient: retryClient.StandardClient(),
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-requested via --insecure flag
	}

	// If retry is disabled, use a plain http.Client without retryablehttp wrapper.
	if c.retryDisabled {
		c.httpClient = &http.Client{
			Transport: transport,
			Timeout:   60 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	return c
}

// NoRetry returns a new client with the same config but retry disabled.
func (c *clientImpl) NoRetry() Client {
	return &clientImpl{
		baseURL:       c.baseURL,
		apiKey:        c.apiKey,
		apiVersion:    c.apiVersion,
		retryDisabled: true,
		insecure:      c.insecure,
		httpClient: &http.Client{
			Transport: c.httpClient.Transport,
			Timeout:   60 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// NewWithHTTPClient creates a client with a custom http.Client (for testing).
func NewWithHTTPClient(baseURL, apiKey, apiVersion string, hc *http.Client) Client {
	return &clientImpl{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		apiVersion: apiVersion,
		httpClient: hc,
	}
}

func (c *clientImpl) Post(ctx context.Context, endpoint string, req any, resp any) error {
	payload, err := c.injectAuth(req)
	if err != nil {
		return errs.Wrap(err, errs.CodeInternalError)
	}

	body, err := c.doPost(ctx, endpoint, payload)
	if err != nil {
		return err
	}
	defer body.Close()

	if err := json.NewDecoder(body).Decode(resp); err != nil {
		return errs.New(errs.CodeInternalError,
			fmt.Sprintf("failed to decode response: %v", err))
	}
	return nil
}

func (c *clientImpl) PostRaw(ctx context.Context, endpoint string, payload map[string]any) (map[string]any, error) {
	payload["ApiKey"] = c.apiKey
	if c.apiVersion != "" {
		payload["ApiVersion"] = c.apiVersion
	}

	body, err := c.doPost(ctx, endpoint, payload)
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var result map[string]any
	dec := json.NewDecoder(body)
	dec.UseNumber()
	if err := dec.Decode(&result); err != nil {
		return nil, errs.New(errs.CodeInternalError,
			fmt.Sprintf("failed to decode response: %v", err))
	}
	return result, nil
}

func (c *clientImpl) injectAuth(req any) (map[string]any, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, fmt.Errorf("failed to convert request to map: %w", err)
	}

	payload["ApiKey"] = c.apiKey
	if c.apiVersion != "" {
		payload["ApiVersion"] = c.apiVersion
	}

	return payload, nil
}

func (c *clientImpl) doPost(ctx context.Context, endpoint string, payload any) (io.ReadCloser, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, errs.Wrap(err, errs.CodeInternalError)
	}

	url := c.baseURL + endpoint
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, errs.Wrap(err, errs.CodeInternalError)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, errs.New(errs.CodeConnectionError,
			fmt.Sprintf("failed to connect to %s: %v", c.baseURL, err)).
			WithSuggestion("Check that the Pleasanter server is running and the URL is correct")
	}

	switch {
	case resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden:
		resp.Body.Close()
		return nil, errs.New(errs.CodeAuthError,
			fmt.Sprintf("authentication failed (HTTP %d)", resp.StatusCode)).
			WithSuggestion("Check your API key with 'plsnt config list'")
	case resp.StatusCode >= 300 && resp.StatusCode < 400:
		// Pleasanter returns 302 on success for createsite operations (redirects to /errors page).
		// Try to read body; if it has JSON content, return it. If empty/non-JSON, return a synthetic success response.
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if len(bodyBytes) > 0 {
			var test map[string]any
			if json.Unmarshal(bodyBytes, &test) == nil {
				return io.NopCloser(bytes.NewReader(bodyBytes)), nil
			}
		}

		// 302 with empty/non-JSON body is normal for Pleasanter createsite.
		// Note: The created resource ID is not available in the response or Location header.
		location := resp.Header.Get("Location")
		msg := fmt.Sprintf("Operation completed (HTTP %d redirect)", resp.StatusCode)
		if location != "" {
			msg += fmt.Sprintf(". Redirect to: %s", location)
		}
		msg += ". Note: Created resource ID is not returned by this API. Use 'site search' or check the Pleasanter UI to find the new resource."
		synthetic := fmt.Sprintf(`{"StatusCode":%d,"Message":%q}`, resp.StatusCode, msg)
		return io.NopCloser(strings.NewReader(synthetic)), nil
	case resp.StatusCode >= 500:
		resp.Body.Close()
		return nil, errs.New(errs.CodeServerError,
			fmt.Sprintf("server error (HTTP %d)", resp.StatusCode))
	case resp.StatusCode >= 400:
		resp.Body.Close()
		return nil, errs.New(errs.CodeValidationError,
			fmt.Sprintf("request error (HTTP %d)", resp.StatusCode))
	}

	return resp.Body, nil
}
