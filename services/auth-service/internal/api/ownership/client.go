package ownership

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"csort.ru/auth-service/internal/config"
	"csort.ru/auth-service/internal/logger"
	"csort.ru/auth-service/internal/observability"
	"github.com/gofiber/fiber/v3/client"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
)

const (
	transferTimeout = 15 * time.Second

	classificationOwnershipTransferPath = "/api/v1/classifications/ownership-transfers"
	analysisOwnershipTransferPath       = "/api/v1/analyses/ownership-transfers"
)

type TokenIssuer func() (string, error)

type targetService struct {
	name string
	url  string
	path string
}

type Client struct {
	services    []targetService
	tokenIssuer TokenIssuer
	fiberClient *client.Client
	log         zerolog.Logger
}

// NewClient builds a client from merge URLs; nil when no downstream services are configured.
func NewClient(cfg *config.MergeConfig, tokenIssuer TokenIssuer) *Client {
	var services []targetService

	if srvURL := strings.TrimSpace(cfg.ClassificationServiceURL); srvURL != "" {
		services = append(services, targetService{
			name: "classification",
			url:  srvURL,
			path: classificationOwnershipTransferPath,
		})
	}

	if srvURL := strings.TrimSpace(cfg.AnalysisServiceURL); srvURL != "" {
		services = append(services, targetService{
			name: "analysis",
			url:  srvURL,
			path: analysisOwnershipTransferPath,
		})
	}

	if len(services) == 0 {
		return nil
	}

	log := logger.GetLogger("ownership.client")

	if tokenIssuer == nil {
		log.Warn().
			Msg("ownership client disabled: token issuer required when downstream URLs are set")
		return nil
	}

	return &Client{
		services:    services,
		tokenIssuer: tokenIssuer,
		fiberClient: client.New().SetTimeout(transferTimeout),
		log:         log,
	}
}

// TransferOwnership POSTs ownership transfer to every configured downstream service (fail-fast).
func (c *Client) TransferOwnership(ctx context.Context, fromUserID, toUserID int32) error {
	token, err := c.tokenIssuer()
	if err != nil {
		return fmt.Errorf("failed to obtain service token for ownership transfer: %w", err)
	}

	payload := OwnershipTransferRequest{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
	}

	for _, svc := range c.services {
		if err := c.postOwnershipTransfer(ctx, svc, token, payload); err != nil {
			c.log.Error().
				Int32("from_user_id", fromUserID).
				Int32("to_user_id", toUserID).
				Str("service", svc.name).
				Err(err).
				Msg("ownership transfer failed")
			return fmt.Errorf("%s service ownership transfer failed: %w", svc.name, err)
		}

		c.log.Info().
			Int32("from_user_id", fromUserID).
			Int32("to_user_id", toUserID).
			Str("service", svc.name).
			Msg("ownership transfer completed")
	}

	return nil
}

func (c *Client) postOwnershipTransfer(
	ctx context.Context,
	svc targetService,
	token string,
	payload OwnershipTransferRequest,
) error {
	endpointURL, err := url.JoinPath(svc.url, svc.path)
	if err != nil {
		return fmt.Errorf("invalid ownership transfer URL: %w", err)
	}

	ctx, sp := observability.StartClientSpan(ctx, "ownership.http.transfer",
		attribute.String("http.request.method", "POST"),
		attribute.String("url.full", endpointURL),
		attribute.String("ownership.target", svc.name),
	)
	if u, perr := url.Parse(endpointURL); perr == nil && u.Host != "" {
		sp.SetAttributes(attribute.String("server.address", u.Host))
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"Authorization": "Bearer " + token,
	}
	observability.InjectTraceHeaders(ctx, func(k, v string) {
		headers[k] = v
	})

	resp, err := c.fiberClient.Post(endpointURL, client.Config{
		Ctx:    ctx,
		Header: headers,
		Body:   payload,
	})
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode()
	}
	if err != nil {
		observability.EndHTTPClientSpan(sp, statusCode, err)
		return err
	}
	defer resp.Close()

	if statusCode < 200 || statusCode >= 300 {
		apiErr := fmt.Errorf(
			"endpoint returned %d: %s",
			statusCode,
			string(resp.Body()),
		)
		observability.EndHTTPClientSpan(sp, statusCode, apiErr)
		return apiErr
	}

	observability.EndHTTPClientSpan(sp, statusCode, nil)
	return nil
}
