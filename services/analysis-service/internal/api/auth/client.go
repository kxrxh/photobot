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

	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/observability"
	"github.com/MicahParks/keyfunc/v2"
	fiberclient "github.com/gofiber/fiber/v3/client"
	jwtk "github.com/golang-jwt/jwt/v5"
	"github.com/valyala/fasthttp"
	"go.opentelemetry.io/otel/attribute"
)

type Config struct {
	BaseURL             string
	ServiceID           string
	ServiceSecret       string
	Timeout             time.Duration
	JWKSRefreshInterval time.Duration
}

type Client struct {
	baseURL             string
	serviceID           string
	serviceSecret       string
	httpClient          *fiberclient.Client
	jwks                *keyfunc.JWKS
	jwksRefreshInterval time.Duration
	token               string
	refreshToken        string
	tokenMu             sync.RWMutex
	mu                  sync.RWMutex
	endJWKSShutdown     sync.Once
}

var authClientLog = logger.GetLogger("api.auth.client")

func recordAuthHTTPError(ctx context.Context, op, fullURL string, statusCode int, err error) {
	if err == nil {
		return
	}
	attrs := []attribute.KeyValue{
		attribute.String("auth.operation", op),
		attribute.String("url.full", fullURL),
	}
	if statusCode > 0 {
		attrs = append(attrs, attribute.Int("http.response.status_code", statusCode))
	}
	if u, perr := url.Parse(fullURL); perr == nil && u.Host != "" {
		attrs = append(attrs, attribute.String("server.address", u.Host))
	}
	observability.RecordError(ctx, err, attrs...)
	authClientLog.Error().
		Err(err).
		Str("op", op).
		Str("url", fullURL).
		Int("status_code", statusCode).
		Msg("auth HTTP request failed")
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.BaseURL == "" {
		return nil, errors.New("auth service URL cannot be empty")
	}

	parsedURL, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid auth service URL: %w", err)
	}
	if parsedURL.Scheme == "" {
		return nil, fmt.Errorf(
			"auth service URL must include a scheme (http:// or https://), got: %s",
			cfg.BaseURL,
		)
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	httpClient := fiberclient.New()
	httpClient.SetTimeout(timeout)

	interval := cfg.JWKSRefreshInterval
	if interval <= 0 {
		interval = time.Hour
	}

	return &Client{
		baseURL:             cfg.BaseURL,
		serviceID:           cfg.ServiceID,
		serviceSecret:       cfg.ServiceSecret,
		httpClient:          httpClient,
		jwksRefreshInterval: interval,
	}, nil
}

func (c *Client) newRequest(ctx context.Context) *fiberclient.Request {
	req := c.httpClient.R().SetContext(ctx)
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

func (c *Client) endpointURL(path string) (string, error) {
	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return "", err
	}
	if baseURL.Path != "" && baseURL.Path != "/" {
		baseURL.Path = strings.TrimSuffix(baseURL.Path, "/") + path
	} else {
		baseURL.Path = path
	}
	return baseURL.String(), nil
}

func (c *Client) Start(ctx context.Context) error {
	if err := c.initKeyfunc(); err != nil {
		return err
	}

	if err := c.RefreshToken(ctx); err != nil {
		return fmt.Errorf("failed to get initial service token: %w", err)
	}

	go func() {
		ticker := time.NewTicker(45 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = c.RefreshToken(ctx)
			}
		}
	}()

	return nil
}

func (c *Client) EndJWKSBackground() {
	c.endJWKSShutdown.Do(func() {
		c.mu.RLock()
		k := c.jwks
		c.mu.RUnlock()
		if k != nil {
			k.EndBackground()
		}
	})
}

func (c *Client) RefreshToken(ctx context.Context) error {
	fullURL, err := c.endpointURL("/auth/login")
	if err != nil {
		return err
	}

	resp, err := c.newRequest(ctx).
		SetHeader("X-Grant-Type", "client_credentials").
		SetJSON(map[string]string{
			"service_id":     c.serviceID,
			"service_secret": c.serviceSecret,
		}).
		Post(fullURL)
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode()
	}
	if err != nil {
		recordAuthHTTPError(ctx, "login", fullURL, statusCode, err)
		return err
	}
	defer resp.Close()

	if resp.StatusCode() != fasthttp.StatusOK {
		bodyBytes := resp.Body()

		var errResp struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(bodyBytes, &errResp)
		apiErr := fmt.Errorf(
			"auth service returned status %d for token refresh: %s",
			resp.StatusCode(),
			errResp.Error,
		)
		recordAuthHTTPError(ctx, "login", fullURL, statusCode, apiErr)
		return apiErr
	}

	var result struct {
		Result struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"result"`
	}
	bodyBytes := resp.Body()
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		recordAuthHTTPError(ctx, "login", fullURL, statusCode, err)
		return err
	}

	c.tokenMu.Lock()
	c.token = result.Result.AccessToken
	c.refreshToken = result.Result.RefreshToken
	c.tokenMu.Unlock()

	return nil
}

func (c *Client) login(ctx context.Context) error {
	return c.RefreshToken(ctx)
}

func (c *Client) ObtainTokens(ctx context.Context) (accessToken, refreshToken string, err error) {
	if err := c.login(ctx); err != nil {
		return "", "", err
	}
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.token, c.refreshToken, nil
}

func (c *Client) RefreshWithRefreshToken(
	ctx context.Context,
	refreshToken string,
) (accessToken, newRefreshToken string, err error) {
	fullURL, err := c.endpointURL("/auth/refresh")
	if err != nil {
		return "", "", err
	}

	resp, err := c.newRequest(ctx).
		SetJSON(map[string]string{"refresh_token": refreshToken}).
		Post(fullURL)
	statusCode := 0
	if resp != nil {
		statusCode = resp.StatusCode()
	}
	if err != nil {
		recordAuthHTTPError(ctx, "refresh", fullURL, statusCode, err)
		return "", "", err
	}
	defer resp.Close()
	if resp.StatusCode() != fasthttp.StatusOK {
		bodyBytes := resp.Body()
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(bodyBytes, &errResp)
		apiErr := fmt.Errorf(
			"auth refresh returned status %d: %s",
			resp.StatusCode(),
			errResp.Error.Message,
		)
		recordAuthHTTPError(ctx, "refresh", fullURL, statusCode, apiErr)
		return "", "", apiErr
	}
	bodyBytes := resp.Body()
	var result struct {
		Result struct {
			AccessToken  string `json:"access_token"`
			RefreshToken string `json:"refresh_token"`
		} `json:"result"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		recordAuthHTTPError(ctx, "refresh", fullURL, statusCode, err)
		return "", "", err
	}
	c.tokenMu.Lock()
	c.token = result.Result.AccessToken
	c.refreshToken = result.Result.RefreshToken
	c.tokenMu.Unlock()
	return result.Result.AccessToken, result.Result.RefreshToken, nil
}

func (c *Client) GetToken() string {
	c.tokenMu.RLock()
	defer c.tokenMu.RUnlock()
	return c.token
}

func (c *Client) initKeyfunc() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.jwks != nil {
		return nil
	}
	fullURL, err := c.endpointURL("/auth/.well-known/jwks.json")
	if err != nil {
		return fmt.Errorf("failed to build JWKS URL: %w", err)
	}
	k, err := keyfunc.Get(fullURL, c.keyfuncOptionsForURL(fullURL))
	if err != nil {
		return err
	}
	c.jwks = k
	return nil
}

func (c *Client) keyfuncOptionsForURL(jwksURL string) keyfunc.Options {
	return keyfunc.Options{
		RefreshInterval: c.jwksRefreshInterval,
		RefreshTimeout:  time.Minute,
		RefreshErrorHandler: func(err error) {
			logger.Logger.Warn().Err(err).Str("url", jwksURL).Msg("jwks background refresh failed")
		},
	}
}

func (c *Client) UpdateJWKS(ctx context.Context) error {
	if err := c.initKeyfunc(); err != nil {
		return err
	}

	c.mu.RLock()
	jwks := c.jwks
	c.mu.RUnlock()

	jwksURL := "keyfunc.Refresh"
	if err := jwks.Refresh(ctx, keyfunc.RefreshOptions{}); err != nil {
		recordAuthHTTPError(ctx, "jwks.refresh", jwksURL, 0, err)
		return err
	}
	return nil
}

func (c *Client) getBotToken(
	ctx context.Context,
	fullURL string,
	headers map[string]string,
) (string, error) {
	c.tokenMu.RLock()
	tok := c.token
	c.tokenMu.RUnlock()
	if tok == "" {
		return "", errors.New("auth client token not initialized")
	}

	parseTokenResponse := func(resp *fiberclient.Response) (string, error) {
		defer resp.Close()
		bodyBytes := resp.Body()
		if resp.StatusCode() != fasthttp.StatusOK {
			var errResp struct {
				Error   string `json:"error"`
				Message string `json:"message"`
			}
			_ = json.Unmarshal(bodyBytes, &errResp)
			msg := errResp.Message
			if msg == "" {
				msg = errResp.Error
			}
			if msg == "" {
				msg = string(bodyBytes)
			}
			return "", fmt.Errorf("auth service returned %d: %s", resp.StatusCode(), msg)
		}
		var result struct {
			Token  string `json:"token"`
			Result struct {
				Token string `json:"token"`
			} `json:"result"`
		}
		if err := json.Unmarshal(bodyBytes, &result); err != nil {
			return "", fmt.Errorf("parse response: %w", err)
		}
		token := result.Token
		if token == "" {
			token = result.Result.Token
		}
		if token == "" {
			return "", errors.New("auth service returned empty token")
		}
		return token, nil
	}

	doRequest := func(token string) (string, int, error) {
		req := c.newRequest(ctx).
			SetHeader("Authorization", "Bearer "+token)
		for k, v := range headers {
			req.SetHeader(k, v)
		}
		resp, err := req.Get(fullURL)
		statusCode := 0
		if resp != nil {
			statusCode = resp.StatusCode()
		}
		if err != nil {
			wrapped := fmt.Errorf("request failed: %w", err)
			recordAuthHTTPError(ctx, "bot_token", fullURL, statusCode, wrapped)
			return "", 0, wrapped
		}
		tokenValue, parseErr := parseTokenResponse(resp)
		if parseErr != nil {
			recordAuthHTTPError(ctx, "bot_token", fullURL, statusCode, parseErr)
			return "", statusCode, parseErr
		}
		return tokenValue, statusCode, nil
	}

	tokenValue, statusCode, reqErr := doRequest(tok)
	if reqErr == nil {
		return tokenValue, nil
	}
	if statusCode != fasthttp.StatusUnauthorized {
		return "", reqErr
	}
	if refreshErr := c.RefreshToken(ctx); refreshErr != nil {
		return "", fmt.Errorf("refresh token after unauthorized: %w", refreshErr)
	}
	c.tokenMu.RLock()
	refreshedToken := c.token
	c.tokenMu.RUnlock()
	if refreshedToken == "" {
		return "", errors.New("auth client token not initialized")
	}
	tokenValue, _, retryErr := doRequest(refreshedToken)
	if retryErr != nil {
		return "", retryErr
	}
	return tokenValue, nil
}

func (c *Client) GetBotTokenByNameAndPlatform(
	ctx context.Context,
	name, platform string,
) (string, error) {
	if name == "" {
		return "", errors.New("bot name is required")
	}
	if platform != "telegram" && platform != "max" {
		return "", fmt.Errorf("invalid platform %q, must be telegram or max", platform)
	}

	fullURL, err := c.endpointURL("/bots/token")
	if err != nil {
		return "", fmt.Errorf("failed to build bot token URL: %w", err)
	}
	return c.getBotToken(ctx, fullURL, map[string]string{
		"X-Bot-Name":           name,
		"X-Messenger-Platform": platform,
	})
}

func (c *Client) ValidateToken(
	ctx context.Context,
	tokenStr string,
) (*TokenValidationResponse, error) {
	if resp, err := c.validateTokenWithJWKS(ctx, tokenStr); err == nil {
		return resp, nil
	}

	c.mu.RLock()
	k := c.jwks
	c.mu.RUnlock()
	if k == nil {
		return &TokenValidationResponse{Valid: false}, errors.New("JWKS not initialized")
	}
	if err := k.Refresh(ctx, keyfunc.RefreshOptions{IgnoreRateLimit: true}); err != nil {
		return &TokenValidationResponse{Valid: false}, fmt.Errorf("failed to refresh JWKS: %w", err)
	}
	return c.validateTokenWithJWKS(ctx, tokenStr)
}

type userJWTSignedClaims struct {
	jwtk.RegisteredClaims
	UserID     *int32   `json:"user_id,omitempty"`
	TelegramID *int64   `json:"telegram_id,omitempty"`
	MaxID      *int64   `json:"max_id,omitempty"`
	Roles      []string `json:"roles"`
	Type       string   `json:"type"`
}

func (c *Client) validateTokenWithJWKS(
	ctx context.Context,
	tokenStr string,
) (*TokenValidationResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	jwksK := c.jwks
	c.mu.RUnlock()

	if jwksK == nil {
		return &TokenValidationResponse{Valid: false}, errors.New("JWKS not initialized")
	}

	var claims userJWTSignedClaims
	_, err := jwtk.ParseWithClaims(
		tokenStr,
		&claims,
		jwksK.Keyfunc,
		jwtk.WithValidMethods([]string{jwtk.SigningMethodRS256.Alg()}),
	)
	if err != nil {
		return &TokenValidationResponse{
				Valid: false,
			}, fmt.Errorf(
				"token verification failed: %w",
				err,
			)
	}

	userID := int32(0)
	if claims.UserID != nil {
		userID = *claims.UserID
	}
	return &TokenValidationResponse{
		Valid: true,
		Identity: &Identity{
			UserID:     userID,
			TelegramID: claims.TelegramID,
			MaxID:      claims.MaxID,
			Roles:      claims.Roles,
		},
		Roles: claims.Roles,
	}, nil
}
