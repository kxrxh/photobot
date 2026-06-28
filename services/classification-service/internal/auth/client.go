package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"csort.ru/classification-service/internal/logger"
	"csort.ru/classification-service/internal/observability"
	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/gofiber/fiber/v3"
	fiberclient "github.com/gofiber/fiber/v3/client"
	"go.opentelemetry.io/otel/attribute"
)

type genericApiResponse struct {
	Success bool            `json:"success"`
	Result  json.RawMessage `json:"result"`
	Error   string          `json:"error,omitempty"`
}

type HTTPStatusError struct {
	StatusCode int
	Path       string
	Body       string
}

func (e *HTTPStatusError) Error() string {
	if e == nil {
		return "unexpected status code"
	}
	if e.Path != "" {
		return fmt.Sprintf("unexpected status code: %d (%s)", e.StatusCode, e.Path)
	}
	return fmt.Sprintf("unexpected status code: %d", e.StatusCode)
}

var authLog = logger.GetLogger("auth.Client")

type Client struct {
	baseURL       string
	fiberClient   *fiberclient.Client
	serviceID     string
	serviceSecret string
	jwks          *jose.JSONWebKeySet
	mu            sync.RWMutex
}

func NewClient(authServiceURL string, serviceID string, serviceSecret string) *Client {
	hc := fiberclient.New()
	hc.SetTimeout(10 * time.Second)
	return &Client{
		baseURL:       authServiceURL,
		serviceID:     serviceID,
		serviceSecret: serviceSecret,
		fiberClient:   hc,
	}
}

func (c *Client) newRequest(ctx context.Context) *fiberclient.Request {
	req := c.fiberClient.R().SetContext(ctx)
	if deadline, ok := ctx.Deadline(); ok {
		timeout := time.Until(deadline)
		if timeout <= 0 {
			timeout = time.Second
		}
		req.SetTimeout(timeout)
	}
	observability.InjectTraceHeaders(ctx, func(k, v string) { req.SetHeader(k, v) })
	return req
}

func (c *Client) Start(ctx context.Context) error {
	if err := c.UpdateJWKS(ctx); err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = c.UpdateJWKS(ctx)
			}
		}
	}()

	return nil
}

func (c *Client) UpdateJWKS(ctx context.Context) error {
	u, err := url.JoinPath(c.baseURL, "/auth/.well-known/jwks.json")
	if err != nil {
		return fmt.Errorf("failed to join URL: %w", err)
	}

	ctx, sp := observability.StartClientSpan(ctx, "auth.http.jwks",
		attribute.String("http.request.method", fiber.MethodGet),
		attribute.String("url.full", u),
	)
	if pu, perr := url.Parse(u); perr == nil && pu.Host != "" {
		sp.SetAttributes(attribute.String("server.address", pu.Host))
	}

	resp, err := c.newRequest(ctx).Get(u) //nolint:gosec // G704: URL from config
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode()
	}
	if err != nil {
		observability.EndHTTPClientSpan(sp, statusCode, err)
		return fmt.Errorf("jwks request failed: %w", err)
	}
	defer resp.Close()

	if resp.StatusCode() != fiber.StatusOK {
		apiErr := fmt.Errorf("auth service returned status %d for JWKS", resp.StatusCode())
		observability.EndHTTPClientSpan(sp, resp.StatusCode(), apiErr)
		return apiErr
	}

	var jwks jose.JSONWebKeySet
	if err := json.Unmarshal(resp.Body(), &jwks); err != nil {
		observability.EndHTTPClientSpan(sp, resp.StatusCode(), err)
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	c.mu.Lock()
	c.jwks = &jwks
	c.mu.Unlock()

	observability.EndHTTPClientSpan(sp, resp.StatusCode(), nil)
	return nil
}

func (c *Client) ValidateToken(
	ctx context.Context,
	tokenStr string,
) (*TokenValidationResponse, error) {
	if resp, err := c.validateTokenWithJWKS(tokenStr); err == nil {
		return resp, nil
	}

	if err := c.UpdateJWKS(ctx); err != nil {
		return &TokenValidationResponse{Valid: false}, fmt.Errorf("failed to refresh JWKS: %w", err)
	}

	return c.validateTokenWithJWKS(tokenStr)
}

func (c *Client) validateTokenWithJWKS(tokenStr string) (*TokenValidationResponse, error) {
	c.mu.RLock()
	jwks := c.jwks
	c.mu.RUnlock()

	if jwks == nil {
		return &TokenValidationResponse{Valid: false}, errors.New("JWKS not initialized")
	}

	tok, err := jwt.ParseSigned(tokenStr, []jose.SignatureAlgorithm{jose.RS256})
	if err != nil {
		return &TokenValidationResponse{Valid: false}, fmt.Errorf("failed to parse token: %w", err)
	}

	var claims CustomClaims
	var stdClaims jwt.Claims
	if err := tok.Claims(jwks, &claims, &stdClaims); err != nil {
		return &TokenValidationResponse{
				Valid: false,
			}, fmt.Errorf(
				"token verification failed: %w",
				err,
			)
	}

	if err := stdClaims.Validate(jwt.Expected{
		Time: time.Now(),
	}); err != nil {
		return &TokenValidationResponse{
				Valid: false,
			}, fmt.Errorf(
				"claims validation failed: %w",
				err,
			)
	}

	return &TokenValidationResponse{
		Valid: true,
		Identity: &Identity{
			UserID:     claims.UserID,
			TelegramID: claims.TelegramID,
			Roles:      claims.Roles,
		},
		Roles: claims.Roles,
	}, nil
}

func (c *Client) doRequest(
	ctx context.Context,
	method, path string,
	reqBody, resBody interface{},
	expectedStatusCodes ...int,
) error {
	return c.doRequestWithAuthHeader(
		ctx,
		method,
		path,
		reqBody,
		resBody,
		"",
		"",
		expectedStatusCodes...)
}

func (c *Client) doRequestWithAuthHeader(
	ctx context.Context,
	method, path string,
	reqBody, resBody interface{},
	headerName, authHeaderValue string,
	expectedStatusCodes ...int,
) error {
	parsedPath, err := url.Parse(path)
	if err != nil {
		return fmt.Errorf("failed to parse path: %w", err)
	}

	u, err := url.JoinPath(c.baseURL, parsedPath.Path)
	if err != nil {
		return fmt.Errorf("failed to join URL: %w", err)
	}
	if parsedPath.RawQuery != "" {
		u = u + "?" + parsedPath.RawQuery
	}

	ctx, sp := observability.StartClientSpan(ctx, "auth.http.request",
		attribute.String("http.request.method", method),
		attribute.String("url.full", u),
	)
	if pu, perr := url.Parse(u); perr == nil && pu.Host != "" {
		sp.SetAttributes(attribute.String("server.address", pu.Host))
	}

	req := c.newRequest(ctx).SetHeader("Content-Type", "application/json")
	if headerName != "" && authHeaderValue != "" {
		req.SetHeader(headerName, authHeaderValue)
	}
	if reqBody != nil {
		req.SetJSON(reqBody)
	}

	var resp *fiberclient.Response
	switch strings.ToUpper(method) {
	case fiber.MethodGet:
		resp, err = req.Get(u) //nolint:gosec // G704: URL from config
	case fiber.MethodPost:
		resp, err = req.Post(u) //nolint:gosec // G704: URL from config
	case fiber.MethodPut:
		resp, err = req.Put(u) //nolint:gosec // G704: URL from config
	case fiber.MethodDelete:
		resp, err = req.Delete(u) //nolint:gosec // G704: URL from config
	default:
		resp, err = req.Custom(u, method) //nolint:gosec // G704: URL from config
	}
	if err != nil {
		observability.EndHTTPClientSpan(sp, 0, err)
		authLog.Error().Err(err).Str("path", path).Msg("Request to auth service failed")
		return err
	}
	defer resp.Close()

	statusCode := resp.StatusCode()
	isExpected := len(expectedStatusCodes) == 0
	for _, code := range expectedStatusCodes {
		if statusCode == code {
			isExpected = true
			break
		}
	}

	if !isExpected {
		bodyStr := string(resp.Body())
		if len(bodyStr) > 1024 {
			bodyStr = bodyStr[:1024]
		}
		authLog.Warn().
			Int("status", statusCode).
			Str("path", path).
			Str("body", bodyStr).
			Msg("Unexpected status code from auth service")
		apiErr := &HTTPStatusError{
			StatusCode: statusCode,
			Path:       path,
			Body:       bodyStr,
		}
		observability.EndHTTPClientSpan(sp, statusCode, apiErr)
		return apiErr
	}

	if resBody != nil {
		bodyBytes := resp.Body()

		if len(bodyBytes) == 0 {
			observability.EndHTTPClientSpan(sp, statusCode, nil)
			return nil
		}

		var wrapper genericApiResponse
		if err := json.Unmarshal(
			bodyBytes,
			&wrapper,
		); err == nil &&
			(wrapper.Success || wrapper.Error != "") {
			if wrapper.Success {
				if err := json.Unmarshal(wrapper.Result, resBody); err != nil {
					observability.EndHTTPClientSpan(sp, statusCode, err)
					return fmt.Errorf("failed to parse nested result: %w", err)
				}
			} else {
				apiErr := fmt.Errorf("auth service error: %s", wrapper.Error)
				observability.EndHTTPClientSpan(sp, statusCode, apiErr)
				return apiErr
			}
		} else {
			if err := json.Unmarshal(bodyBytes, resBody); err != nil {
				observability.EndHTTPClientSpan(sp, statusCode, err)
				return fmt.Errorf("failed to parse response body: %w", err)
			}
		}
	}

	observability.EndHTTPClientSpan(sp, statusCode, nil)
	return nil
}

func (c *Client) LoginAsService(ctx context.Context, audience string) (string, string, error) {
	req := map[string]string{
		"service_id":     c.serviceID,
		"service_secret": c.serviceSecret,
		"audience":       audience,
	}

	var resp struct {
		AccessToken  string `json:"access_token"`  //nolint:gosec // G117: API response field
		RefreshToken string `json:"refresh_token"` //nolint:gosec // G117: API response field
	}

	err := c.doRequestWithAuthHeader(
		ctx,
		"POST",
		"/auth/login",
		req,
		&resp,
		"X-Grant-Type",
		"client_credentials",
		fiber.StatusOK,
	)
	if err != nil {
		return "", "", err
	}

	return resp.AccessToken, resp.RefreshToken, nil
}

func (c *Client) RefreshTokens(ctx context.Context, refreshToken string) (string, string, error) {
	req := map[string]string{
		"refresh_token": refreshToken,
	}

	var resp struct {
		AccessToken  string `json:"access_token"`  //nolint:gosec // G117: API response field
		RefreshToken string `json:"refresh_token"` //nolint:gosec // G117: API response field
	}

	err := c.doRequest(ctx, "POST", "/auth/refresh", req, &resp, fiber.StatusOK)
	if err != nil {
		return "", "", err
	}

	return resp.AccessToken, resp.RefreshToken, nil
}
