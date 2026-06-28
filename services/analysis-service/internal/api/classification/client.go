package classification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"csort.ru/analysis-service/internal/api/common"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/observability"
	"github.com/bytedance/sonic"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/attribute"
)

const peerService = "classification-service"

var classificationClientLog = logger.GetLogger("api.classification.client")

type Config struct {
	BaseURL      string
	Timeout      time.Duration
	MaxRetries   int
	RetryDelay   time.Duration
	GetToken     func() string
	RefreshToken func(ctx context.Context) error
}

type Client struct {
	baseURL    string
	httpClient *fasthttp.Client
	config     Config
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("classification BaseURL cannot be empty")
	}
	u, err := url.ParseRequestURI(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid classification BaseURL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("classification BaseURL must have http or https scheme")
	}

	path := strings.TrimRight(u.Path, "/")
	if !strings.HasSuffix(path, "/api/v1") {
		return nil, errors.New("classification BaseURL must include /api/v1")
	}
	if cfg.MaxRetries < 0 {
		return nil, errors.New("MaxRetries cannot be negative")
	}
	if cfg.Timeout < 0 || cfg.RetryDelay < 0 {
		return nil, errors.New("timeout and RetryDelay cannot be negative")
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 2
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 200 * time.Millisecond
	}

	base := strings.TrimRight(cfg.BaseURL, "/")
	return &Client{
		baseURL: base,
		config:  cfg,
		httpClient: &fasthttp.Client{
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			Name:         "classification-client",
		},
	}, nil
}

type envelope struct {
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result"`
}

func (c *Client) GetUserActiveClassification(
	ctx context.Context,
	messengerUserID string,
	platform string,
) (json.RawMessage, error) {
	if messengerUserID == "" {
		return nil, errors.New("messengerUserID cannot be empty")
	}
	normalizedPlatform := strings.TrimSpace(strings.ToLower(platform))
	if normalizedPlatform != "telegram" && normalizedPlatform != "max" {
		return nil, errors.New("platform must be 'telegram' or 'max'")
	}

	endpoint := c.baseURL +
		"/user-active-classifications/" + url.PathEscape(messengerUserID) +
		"?platform=" + url.QueryEscape(normalizedPlatform)

	statusCode, body, err := c.do(
		ctx,
		"get_user_active_classification",
		http.MethodGet,
		endpoint,
		true,
	)
	if err != nil {
		return nil, err
	}

	switch statusCode {
	case http.StatusNoContent:
		return nil, nil
	case http.StatusOK:
		var env envelope
		if err := sonic.Unmarshal(body, &env); err != nil {
			apiErr := common.NewHTTPError(
				peerService, http.MethodGet, endpoint, statusCode, body,
				"failed to unmarshal classification response envelope", err,
			)
			observability.UpstreamHTTPError(
				ctx, classificationClientLog, peerService, "get_user_active_classification",
				http.MethodGet, endpoint, statusCode, body, apiErr,
			)
			return nil, apiErr
		}
		if !env.Success {
			apiErr := common.NewHTTPError(
				peerService, http.MethodGet, endpoint, statusCode, body,
				"classification response success=false", nil,
			)
			observability.UpstreamHTTPError(
				ctx, classificationClientLog, peerService, "get_user_active_classification",
				http.MethodGet, endpoint, statusCode, body, apiErr,
			)
			return nil, apiErr
		}
		return env.Result, nil
	default:
		apiErr := common.NewHTTPError(
			peerService, http.MethodGet, endpoint, statusCode, body,
			fmt.Sprintf("unexpected status code: %d", statusCode), nil,
		)
		observability.UpstreamHTTPError(
			ctx, classificationClientLog, peerService, "get_user_active_classification",
			http.MethodGet, endpoint, statusCode, body, apiErr,
		)
		return nil, apiErr
	}
}

func (c *Client) do(
	ctx context.Context,
	operation, method, endpoint string,
	authenticate bool,
) (statusCode int, body []byte, err error) {
	ctx, span := observability.StartClientSpan(ctx, "classification."+operation,
		attribute.String("peer.service", peerService),
		attribute.String("http.request.method", method),
		attribute.String("url.full", endpoint),
		attribute.String("upstream.operation", operation),
	)
	defer func() {
		observability.EndHTTPClientSpan(span, statusCode, err)
	}()

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.SetRequestURI(endpoint)
	req.Header.SetMethod(method)
	if authenticate {
		if c.config.GetToken == nil {
			return 0, nil, errors.New("GetToken is nil")
		}
		token := strings.TrimSpace(c.config.GetToken())
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	var lastErr error
	refreshed := false
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return 0, nil, ctx.Err()
			case <-time.After(c.config.RetryDelay):
			}
		}

		observability.InjectTraceFasthttp(ctx, req)

		lastErr = c.httpClient.Do(req, resp)
		if lastErr != nil {
			continue
		}

		statusCode = resp.StatusCode()
		body = append([]byte(nil), resp.Body()...)

		if statusCode == http.StatusUnauthorized && authenticate && c.config.RefreshToken != nil &&
			!refreshed {
			if refreshErr := c.config.RefreshToken(ctx); refreshErr != nil {
				apiErr := common.NewHTTPError(
					peerService, method, endpoint, statusCode, body,
					"unauthorized; token refresh failed", refreshErr,
				)
				observability.UpstreamHTTPError(
					ctx, classificationClientLog, peerService, operation,
					method, endpoint, statusCode, body, apiErr,
				)
				return statusCode, body, apiErr
			}
			token := strings.TrimSpace(c.config.GetToken())
			req.Header.Set("Authorization", "Bearer "+token)
			refreshed = true
			continue
		}

		return statusCode, body, nil
	}

	transportErr := fmt.Errorf("classification HTTP request failed: %w", lastErr)
	observability.UpstreamHTTPError(
		ctx, classificationClientLog, peerService, operation,
		method, endpoint, 0, nil, transportErr,
	)
	return 0, nil, transportErr
}
