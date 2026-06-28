package auth

import (
	"context"
	"sync"
	"time"

	"csort.ru/classification-service/internal/logger"
)

var tokenLog = logger.GetLogger("auth.token_manager")

type TokenManager struct {
	authClient   *Client
	audience     string
	accessToken  string
	refreshToken string
	mu           sync.RWMutex
}

func NewTokenManager(authClient *Client, audience string) *TokenManager {
	return &TokenManager{
		authClient: authClient,
		audience:   audience,
	}
}

func (m *TokenManager) Start(ctx context.Context) error {
	if err := m.refresh(ctx); err != nil {
		tokenLog.Error().Err(err).Str("audience", m.audience).Msg("Failed initial token fetch")
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := m.refresh(ctx); err != nil {
					tokenLog.Error().
						Err(err).
						Str("audience", m.audience).
						Msg("Failed to refresh token")
				}
			}
		}
	}()

	return nil
}

func (m *TokenManager) refresh(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var accessToken, refreshToken string
	var err error

	if m.refreshToken != "" {
		accessToken, refreshToken, err = m.authClient.RefreshTokens(ctx, m.refreshToken)
		if err != nil {
			tokenLog.Warn().
				Err(err).
				Str("audience", m.audience).
				Msg("Failed to refresh tokens, attempting full login")
			m.refreshToken = "" // Clear invalid refresh token
		}
	}

	if m.refreshToken == "" {
		accessToken, refreshToken, err = m.authClient.LoginAsService(ctx, m.audience)
		if err != nil {
			return err
		}
	}

	m.accessToken = accessToken
	m.refreshToken = refreshToken
	tokenLog.Info().Str("audience", m.audience).Msg("Successfully refreshed service token")
	return nil
}

func (m *TokenManager) GetToken() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.accessToken
}

func (m *TokenManager) RefreshToken(ctx context.Context) error {
	return m.refresh(ctx)
}
