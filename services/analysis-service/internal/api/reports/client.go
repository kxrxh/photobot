package reports

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"csort.ru/analysis-service/internal/api/common"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/observability"
	"csort.ru/analysis-service/internal/report"
	"github.com/bytedance/sonic"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/attribute"
)

const reportsPeerService = "reports-service"

var reportsClientLog = logger.GetLogger("api.reports.client")

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
		return nil, errors.New("reports BaseURL cannot be empty")
	}

	u, err := url.ParseRequestURI(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid reports BaseURL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("reports BaseURL must have http or https scheme")
	}
	if cfg.MaxRetries < 0 {
		return nil, errors.New("MaxRetries cannot be negative")
	}
	if cfg.Timeout < 0 || cfg.RetryDelay < 0 {
		return nil, errors.New("timeout and RetryDelay cannot be negative")
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = time.Second
	}

	base := strings.TrimRight(cfg.BaseURL, "/")

	return &Client{
		baseURL: base,
		config:  cfg,
		httpClient: &fasthttp.Client{
			ReadTimeout:  cfg.Timeout,
			WriteTimeout: cfg.Timeout,
			Name:         "reports-client",
		},
	}, nil
}

func (c *Client) GenerateReport(
	ctx context.Context,
	analysisID string,
) (*report.ReportResponse, error) {
	if analysisID == "" {
		return nil, errors.New("analysisID cannot be empty")
	}

	endpoint := c.endpoint("api", "reports", analysisID)

	statusCode, body, err := c.doFasthttp(ctx, http.MethodPut, endpoint, "", nil, true)
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		apiErr := common.NewHTTPError(
			reportsPeerService,
			http.MethodPut,
			endpoint,
			statusCode,
			body,
			fmt.Sprintf("unexpected status code: %d", statusCode),
			nil,
		)
		observability.UpstreamHTTPError(
			ctx, reportsClientLog, reportsPeerService, "generate_report",
			http.MethodPut, endpoint, statusCode, body, apiErr,
		)
		return nil, apiErr
	}

	var out report.ReportResponse
	if err := sonic.Unmarshal(body, &out); err != nil {
		return nil, common.NewHTTPError(
			reportsPeerService, http.MethodPut, endpoint, statusCode, body,
			"failed to unmarshal success response", err,
		)
	}

	return &out, nil
}

func (c *Client) DownloadReportToWriter(
	ctx context.Context,
	analysisID, fileType string,
	dst io.Writer,
) (n int64, err error) {
	if analysisID == "" {
		return 0, errors.New("analysisID cannot be empty")
	}
	if fileType == "" {
		return 0, errors.New("fileType cannot be empty")
	}
	if dst == nil {
		return 0, errors.New("dst cannot be nil")
	}

	body, err := c.downloadReportContent(ctx, analysisID, fileType)
	if err != nil {
		return 0, err
	}

	n, err = io.Copy(dst, bytes.NewReader(body))
	if err != nil {
		return n, err
	}
	reportsClientLog.Debug().
		Str("analysis_id", analysisID).
		Str("format", fileType).
		Int64("bytes", n).
		Msg("reports download: completed (buffered)")
	return n, nil
}

type presignedDownloadEnvelope struct {
	URL              string `json:"url"`
	ExpiresInSeconds int64  `json:"expiresInSeconds"`
}

func (c *Client) downloadReportContent(
	ctx context.Context,
	analysisID, fileType string,
) ([]byte, error) {
	meta := c.endpoint("api", "reports", analysisID, "download-url") +
		"?format=" + url.QueryEscape(fileType)

	metaStatus, metaBody, err := c.doFasthttp(ctx, http.MethodGet, meta, "", nil, true)
	if err != nil {
		return nil, err
	}
	if metaStatus != http.StatusOK {
		apiErr := common.NewHTTPError(
			reportsPeerService, http.MethodGet, meta, metaStatus, metaBody,
			fmt.Sprintf("failed to fetch report download URL (%d)", metaStatus), nil,
		)
		observability.UpstreamHTTPError(
			ctx, reportsClientLog, reportsPeerService, "download_report_meta",
			http.MethodGet, meta, metaStatus, metaBody, apiErr,
		)
		return nil, apiErr
	}
	var env presignedDownloadEnvelope
	if err := sonic.Unmarshal(metaBody, &env); err != nil {
		return nil, common.NewHTTPError(
			reportsPeerService, http.MethodGet, meta, metaStatus, metaBody,
			"failed to unmarshal reports download-url response", err,
		)
	}
	fetchURL := strings.TrimSpace(env.URL)
	if fetchURL == "" {
		return nil, common.NewHTTPError(
			reportsPeerService, http.MethodGet, meta, http.StatusBadGateway, metaBody,
			"reports download-url returned an empty URL", nil,
		)
	}

	blobStatus, blob, err := c.doFasthttp(ctx, http.MethodGet, fetchURL, "", nil, false)
	if err != nil {
		return nil, err
	}
	if blobStatus != http.StatusOK {
		apiErr := common.NewHTTPError(
			reportsPeerService, http.MethodGet, fetchURL, blobStatus, blob,
			fmt.Sprintf("failed to GET presigned report object (%d)", blobStatus), nil,
		)
		observability.UpstreamHTTPError(
			ctx, reportsClientLog, reportsPeerService, "download_report_blob",
			http.MethodGet, fetchURL, blobStatus, blob, apiErr,
		)
		return nil, apiErr
	}
	return blob, nil
}

func (c *Client) endpoint(parts ...string) string {
	clean := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.Trim(p, "/")
		if p != "" {
			clean = append(clean, p)
		}
	}
	if len(clean) == 0 {
		return c.baseURL
	}
	return c.baseURL + "/" + strings.Join(clean, "/")
}

func (c *Client) bearerToken() string {
	if c.config.GetToken == nil {
		return ""
	}
	return strings.TrimSpace(c.config.GetToken())
}

func (c *Client) setAuthHeaderFast(req *fasthttp.Request) {
	if token := c.bearerToken(); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func (c *Client) refreshToken(ctx context.Context) error {
	if c.config.RefreshToken == nil {
		return nil
	}
	return c.config.RefreshToken(ctx)
}

func (c *Client) doFasthttp(
	ctx context.Context,
	method, endpoint, contentType string,
	body []byte,
	sendBearer bool,
) (int, []byte, error) {
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(endpoint)
	req.Header.SetMethod(method)
	if contentType != "" {
		req.Header.SetContentType(contentType)
	}
	if len(body) > 0 {
		req.SetBodyRaw(body)
	}
	if sendBearer {
		c.setAuthHeaderFast(req)
	}

	var (
		did401Retry bool
		lastErr     error
	)

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			if err := sleepWithContext(ctx, c.config.RetryDelay); err != nil {
				return 0, nil, err
			}
		}

		resp.Reset()

		observability.InjectTraceFasthttp(ctx, req)

		lastErr = c.httpClient.Do(req, resp)
		if lastErr != nil {
			reportsClientLog.Warn().
				Err(lastErr).
				Str("method", method).
				Str("endpoint", endpoint).
				Int("attempt", attempt+1).
				Msg("reports HTTP: transport error (may retry)")
			continue
		}

		status := resp.StatusCode()

		if sendBearer && status == http.StatusUnauthorized && c.config.RefreshToken != nil &&
			!did401Retry {
			if err := c.refreshToken(ctx); err != nil {
				respBody := copyFastBody(resp)
				apiErr := common.NewHTTPError(
					reportsPeerService, method, endpoint, status, respBody,
					"unauthorized and token refresh failed", err,
				)
				observability.UpstreamHTTPError(
					ctx, reportsClientLog, reportsPeerService, "token_refresh",
					method, endpoint, status, respBody, apiErr,
				)
				return status, respBody, apiErr
			}

			c.setAuthHeaderFast(req)
			did401Retry = true
			continue
		}

		if status >= 500 && attempt < c.config.MaxRetries {
			reportsClientLog.Warn().
				Str("endpoint", endpoint).
				Int("status", status).
				Int("attempt", attempt+1).
				Msg("reports HTTP: upstream 5xx, retrying")
			continue
		}

		respBody := copyFastBody(resp)

		if status >= 400 {
			apiErr := common.NewHTTPError(
				reportsPeerService,
				method,
				endpoint,
				status,
				respBody,
				"API returned error",
				nil,
			)
			observability.UpstreamHTTPError(
				ctx, reportsClientLog, reportsPeerService, "http_request",
				method, endpoint, status, respBody, apiErr,
			)
			return status, respBody, apiErr
		}

		return status, respBody, nil
	}

	if lastErr != nil {
		observability.RecordError(ctx, lastErr,
			attribute.String("http.request.method", method),
			attribute.String("url.full", endpoint),
			attribute.String("reports.failure", "max_retries_transport"),
		)
		reportsClientLog.Error().
			Err(lastErr).
			Str("method", method).
			Str("endpoint", endpoint).
			Msg("reports HTTP: max retries reached for network errors")
		apiErr := common.NewHTTPError(
			reportsPeerService, method, endpoint, 0, nil,
			"max retries reached for network request", lastErr,
		)
		observability.UpstreamHTTPError(
			ctx, reportsClientLog, reportsPeerService, "http_request",
			method, endpoint, 0, nil, apiErr,
		)
		return 0, nil, apiErr
	}

	unknown := errors.New("request failed for unknown reason")
	apiErr := common.NewHTTPError(
		reportsPeerService, method, endpoint, 0, nil,
		"request failed for unknown reason", unknown,
	)
	observability.UpstreamHTTPError(
		ctx, reportsClientLog, reportsPeerService, "http_request",
		method, endpoint, 0, nil, apiErr,
	)
	return 0, nil, apiErr
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

func copyFastBody(resp *fasthttp.Response) []byte {
	b := resp.Body()
	if len(b) == 0 {
		return nil
	}
	out := make([]byte, len(b))
	copy(out, b)
	return out
}
