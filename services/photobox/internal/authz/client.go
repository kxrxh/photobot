package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"csort.ru/coffeebot/internal/observability"
	"github.com/bytedance/sonic"
	"github.com/go-jose/go-jose/v4"
	"github.com/rs/zerolog"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/sync/singleflight"
)

type genericApiResponse struct {
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result"`
	Error   string          `json:"error,omitempty"`
}

// HTTPClient is the shared fasthttp pool used for identity service requests.
var HTTPClient = &fasthttp.Client{
	MaxConnsPerHost: 512,
	ReadTimeout:     10 * time.Second,
	WriteTimeout:    10 * time.Second,
}

// Client is an HTTP client for the identity service (JWKS, users, roles).
type Client struct {
	baseURL          string
	client           *fasthttp.Client
	httpClient       *http.Client
	jwks             *jose.JSONWebKeySet
	mu               sync.RWMutex
	jwksCancel       context.CancelFunc
	jwksRefreshGroup singleflight.Group
	log              zerolog.Logger
}

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout:   10 * time.Second,
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}

// NewClient builds a Client for the given identity base URL.
func NewClient(identityServiceURL string, log zerolog.Logger) *Client {
	return &Client{
		baseURL:    identityServiceURL,
		client:     HTTPClient,
		httpClient: newHTTPClient(),
		log:        log,
	}
}

func (c *Client) Start(ctx context.Context) error {
	if err := c.UpdateJWKS(ctx); err != nil {
		return err
	}

	runCtx, cancel := context.WithCancel(ctx)
	c.mu.Lock()
	if c.jwksCancel != nil {
		c.jwksCancel()
	}
	c.jwksCancel = cancel
	c.mu.Unlock()

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				_ = c.UpdateJWKS(runCtx)
			}
		}
	}()

	return nil
}

// Stop cancels the hourly JWKS refresh goroutine (idempotent).
func (c *Client) Stop() {
	c.mu.Lock()
	cancel := c.jwksCancel
	c.jwksCancel = nil
	c.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// UpdateJWKS fetches the latest JWKS from the auth service.
func (c *Client) UpdateJWKS(ctx context.Context) error {
	endpoint, err := url.JoinPath(c.baseURL, "/auth/.well-known/jwks.json")
	if err != nil {
		return fmt.Errorf("failed to join URL: %w", err)
	}

	ctx, sp := observability.StartClientSpan(ctx, "authz.http.jwks",
		attribute.String("http.request.method", http.MethodGet),
		attribute.String("url.full", endpoint),
	)
	if pu, perr := url.Parse(endpoint); perr == nil && pu.Host != "" {
		sp.SetAttributes(attribute.String("server.address", pu.Host))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		observability.EndHTTPClientSpan(sp, 0, err)
		return fmt.Errorf("failed to create request: %w", err)
	}
	observability.InjectTraceHTTPRequest(ctx, req)

	resp, err := c.httpClient.Do(req) //nolint:gosec // G704: JWKS URL from config, not user input
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode
	}
	if err != nil {
		observability.EndHTTPClientSpan(sp, statusCode, err)
		return fmt.Errorf("jwks request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		apiErr := fmt.Errorf("auth service returned status %d for JWKS", resp.StatusCode)
		observability.EndHTTPClientSpan(sp, resp.StatusCode, apiErr)
		return apiErr
	}

	var jwks jose.JSONWebKeySet
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		observability.EndHTTPClientSpan(sp, resp.StatusCode, err)
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	c.mu.Lock()
	c.jwks = &jwks
	c.mu.Unlock()

	observability.EndHTTPClientSpan(sp, resp.StatusCode, nil)
	return nil
}

func (c *Client) doRequest(
	ctx context.Context,
	method, path string,
	reqBody, resBody interface{},
	expectedStatusCodes ...int,
) (err error) {
	fullURL := c.baseURL + path
	ctx, sp := observability.StartClientSpan(ctx, "authz.http.request",
		attribute.String("http.request.method", method),
		attribute.String("url.full", fullURL),
	)
	if u, perr := url.Parse(fullURL); perr == nil && u.Host != "" {
		sp.SetAttributes(attribute.String("server.address", u.Host))
	}

	statusCode := 0
	defer func() { observability.EndHTTPClientSpan(sp, statusCode, err) }()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(fullURL)
	req.Header.SetMethod(method)
	req.Header.SetContentType("application/json")

	observability.InjectTraceFasthttp(ctx, req)

	if reqBody != nil {
		body, mErr := sonic.Marshal(reqBody)
		if mErr != nil {
			err = fmt.Errorf("failed to marshal request body: %w", mErr)
			return err
		}
		req.SetBody(body)
	}

	if deadline, ok := ctx.Deadline(); ok {
		err = c.client.DoTimeout(req, resp, time.Until(deadline))
	} else {
		err = c.client.Do(req, resp)
	}
	if err != nil {
		c.log.Error().Err(err).Str("path", path).Msg("Request to identity service failed")
		return err
	}

	statusCode = resp.StatusCode()

	if len(expectedStatusCodes) > 0 {
		isExpected := false
		for _, code := range expectedStatusCodes {
			if statusCode == code {
				isExpected = true
				break
			}
		}
		if !isExpected {
			err = fmt.Errorf(
				"identity service returned status %d, expected one of %v",
				statusCode,
				expectedStatusCodes,
			)
			c.log.Warn().
				Err(err).
				Str("path", path).
				Msg("Unexpected status code from identity service")
			return err
		}
	}

	if resBody != nil && len(resp.Body()) > 0 {
		var wrapper genericApiResponse
		// Identity may respond with {success,result} or a raw JSON body.
		if uErr := sonic.Unmarshal(resp.Body(), &wrapper); uErr == nil && wrapper.Success {
			if uErr := sonic.Unmarshal(wrapper.Result, resBody); uErr != nil {
				c.log.Error().
					Err(uErr).
					Str("path", path).
					Str("body", string(wrapper.Result)).
					Msg("Failed to parse nested result from identity service")
				err = fmt.Errorf("failed to parse nested result from identity service: %w", uErr)
				return err
			}
		} else {
			if uErr := sonic.Unmarshal(resp.Body(), resBody); uErr != nil {
				c.log.Error().
					Err(uErr).
					Str("path", path).
					Str("body", string(resp.Body())).
					Msg("Failed to parse response body from identity service")
				err = fmt.Errorf("failed to parse response body: %w", uErr)
				return err
			}
		}
	}

	return nil
}
