package correlation

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"csort.ru/classification-service/internal/logger"
	"csort.ru/classification-service/internal/observability"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
	fiberclient "github.com/gofiber/fiber/v3/client"
	"go.opentelemetry.io/otel/attribute"
)

var correlationLog = logger.GetLogger("services.correlation")

var ErrUnauthorized = errors.New("unauthorized: token expired or invalid")

type Client struct {
	baseURL string
	client  *fiberclient.Client
}

func NewClient(baseURL string) *Client {
	hc := fiberclient.New()
	hc.SetTimeout(30 * time.Second)
	return &Client{
		baseURL: baseURL,
		client:  hc,
	}
}

func (c *Client) CalculateCorrelation(
	ctx context.Context,
	req *CorrelationRequest,
	token string,
) ([]CorrelationWithTest, error) {
	endpoint := c.baseURL + "/correlation-api/"

	body, err := sonic.Marshal(req)
	if err != nil {
		correlationLog.Error().Err(err).Msg("Failed to marshal correlation request")
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, sp := observability.StartClientSpan(ctx, "correlation.http.request",
		attribute.String("http.request.method", fiber.MethodPost),
		attribute.String("url.full", endpoint),
	)
	if u, perr := url.Parse(endpoint); perr == nil && u.Host != "" {
		sp.SetAttributes(attribute.String("server.address", u.Host))
	}

	headers := map[string]string{"Content-Type": "application/json"}
	if token != "" {
		headers["Authorization"] = "Bearer " + token
	}
	observability.InjectTraceHeaders(ctx, func(k, v string) { headers[k] = v })

	httpResp, err := c.client.Post(
		endpoint,
		fiberclient.Config{ //nolint:gosec // G704: URL from config
			Ctx:    ctx,
			Header: headers,
			Body:   body,
		},
	)
	statusCode := 0
	if httpResp != nil {
		statusCode = httpResp.StatusCode()
	}
	if err != nil {
		correlationLog.Error().Err(err).Msg("Request to correlation service failed")
		observability.EndHTTPClientSpan(sp, statusCode, err)
		return nil, err
	}
	defer httpResp.Close()

	if httpResp.StatusCode() != fiber.StatusOK {
		bodyBytes := httpResp.Body()
		correlationLog.Error().
			Int("status", httpResp.StatusCode()).
			Str("body", string(bodyBytes)).
			Msg("Correlation service returned error")

		if httpResp.StatusCode() == fiber.StatusUnauthorized {
			observability.EndHTTPClientSpan(sp, httpResp.StatusCode(), nil)
			return nil, ErrUnauthorized
		}

		apiErr := fmt.Errorf(
			"correlation service returned status %d: %s",
			httpResp.StatusCode(),
			string(bodyBytes),
		)
		observability.EndHTTPClientSpan(sp, httpResp.StatusCode(), apiErr)
		return nil, apiErr
	}

	bodyBytes := httpResp.Body()

	var result []CorrelationWithTest
	if err := sonic.Unmarshal(bodyBytes, &result); err != nil {
		correlationLog.Error().
			Err(err).
			Str("body", string(bodyBytes)).
			Msg("Failed to unmarshal correlation response")
		observability.EndHTTPClientSpan(sp, httpResp.StatusCode(), err)
		return nil, fmt.Errorf("failed to unmarshal correlation response: %w", err)
	}

	observability.EndHTTPClientSpan(sp, httpResp.StatusCode(), nil)
	return result, nil
}
