package worker

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"csort.ru/analysis-service/internal/api/common"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/observability"
	"github.com/bytedance/sonic"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/attribute"
)

const peerService = "kalibr-worker"

type Config struct {
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
	RetryDelay time.Duration
}

type Client struct {
	defaultBaseURL string
	config         Config
	httpClient     *fasthttp.Client
}

func NewClient(cfg Config) *Client {
	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 2
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 200 * time.Millisecond
	}

	return &Client{
		defaultBaseURL: cfg.BaseURL,
		config:         cfg,
		httpClient: &fasthttp.Client{
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
		},
	}
}

var workerClientLog = logger.GetLogger("api.worker.client")

func (c *Client) baseURL(context.Context) string {
	return c.defaultBaseURL
}

func (c *Client) RecalculateAnalysis(
	ctx context.Context,
	tempID string,
	excludedObjects []string,
) (string, error) {
	if tempID == "" {
		return "", errors.New("request ID cannot be empty")
	}

	if excludedObjects == nil {
		excludedObjects = []string{}
	}

	payload := map[string]any{
		"id":                  tempID,
		"excluded_object_ids": excludedObjects,
	}

	bodyBytes, err := sonic.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	endpoint, err := url.JoinPath(c.baseURL(ctx), "recalculate")
	if err != nil {
		return "", fmt.Errorf("failed to build endpoint: %w", err)
	}

	statusCode, responseBody, err := c.makeRequest(
		ctx,
		"recalculate",
		"POST",
		endpoint,
		bodyBytes,
		"application/json",
	)
	if err != nil {
		return "", fmt.Errorf("failed to recalculate analysis for ID %s: %w", tempID, err)
	}

	if statusCode != http.StatusOK {
		apiErr := common.NewHTTPError(
			peerService,
			"POST",
			endpoint,
			statusCode,
			responseBody,
			fmt.Sprintf("recalculate API returned status %d for ID %s", statusCode, tempID),
			nil,
		)
		observability.UpstreamHTTPError(
			ctx, workerClientLog, peerService, "recalculate",
			"POST", endpoint, statusCode, responseBody, apiErr,
		)
		return "", apiErr
	}

	var apiResp struct {
		Response string `json:"Response"`
	}
	if err := sonic.Unmarshal(responseBody, &apiResp); err != nil {
		return "", errors.New("invalid response format from recalculate API")
	}

	return apiResp.Response, nil
}

func (c *Client) MergeAnalyses(ctx context.Context, payload map[string]any) ([]byte, error) {
	bodyBytes, err := sonic.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	endpoint, err := url.JoinPath(c.baseURL(ctx), "merge")
	if err != nil {
		return nil, fmt.Errorf("failed to build endpoint: %w", err)
	}

	statusCode, responseBody, err := c.makeRequest(
		ctx,
		"merge",
		"POST",
		endpoint,
		bodyBytes,
		"application/json",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to merge analyses: %w", err)
	}

	if statusCode != http.StatusOK {
		apiErr := common.NewHTTPError(
			peerService,
			"POST",
			endpoint,
			statusCode,
			responseBody,
			fmt.Sprintf("merge API returned status %d", statusCode),
			nil,
		)
		observability.UpstreamHTTPError(
			ctx, workerClientLog, peerService, "merge",
			"POST", endpoint, statusCode, responseBody, apiErr,
		)
		return nil, apiErr
	}

	return responseBody, nil
}

func (c *Client) makeRequest(
	ctx context.Context,
	operation, method, rawURL string,
	body []byte,
	contentType string,
) (statusCode int, responseBody []byte, err error) {
	ctx, span := observability.StartClientSpan(ctx, "worker."+operation,
		attribute.String("peer.service", peerService),
		attribute.String("http.request.method", method),
		attribute.String("url.full", rawURL),
		attribute.String("upstream.operation", operation),
	)
	defer func() {
		observability.EndHTTPClientSpan(span, statusCode, err)
	}()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(rawURL)
	req.Header.SetMethod(method)
	if contentType != "" {
		req.Header.SetContentType(contentType)
	}
	if body != nil {
		req.SetBody(body)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	var lastErr error
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			if sleepErr := sleepWithContext(ctx, c.config.RetryDelay); sleepErr != nil {
				return 0, nil, sleepErr
			}
		}

		resp.Reset()
		observability.InjectTraceFasthttp(ctx, req)

		lastErr = c.httpClient.Do(req, resp)
		if lastErr != nil {
			workerClientLog.Warn().
				Err(lastErr).
				Str("operation", operation).
				Str("method", method).
				Str("endpoint", rawURL).
				Int("attempt", attempt+1).
				Msg("kalibr worker HTTP: transport error (may retry)")
			continue
		}

		statusCode = resp.StatusCode()
		responseBody = make([]byte, len(resp.Body()))
		copy(responseBody, resp.Body())

		if statusCode >= 500 && attempt < c.config.MaxRetries {
			workerClientLog.Warn().
				Str("operation", operation).
				Str("endpoint", rawURL).
				Int("status", statusCode).
				Int("attempt", attempt+1).
				Msg("kalibr worker HTTP: upstream 5xx, retrying")
			continue
		}

		return statusCode, responseBody, nil
	}

	if lastErr != nil {
		observability.UpstreamHTTPError(
			ctx, workerClientLog, peerService, operation,
			method, rawURL, 0, nil, lastErr,
		)
		return 0, nil, fmt.Errorf("HTTP request failed: %w", lastErr)
	}

	return statusCode, responseBody, nil
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
