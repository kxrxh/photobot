package common

import (
	"context"
	"errors"
	"sync"
	"time"
)

var (
	ErrTokenManagerNoObtainTokens = errors.New(
		"token manager: ObtainTokens is required",
	)
	ErrTokenManagerNoRefreshWithRefresh = errors.New(
		"token manager: RefreshWithRefreshToken is required",
	)
)

type TokenManager struct {
	obtainTokens          func(context.Context) (string, string, error)
	refreshWithRefreshTok func(context.Context, string) (string, string, error)
	interval              time.Duration

	mu           sync.RWMutex
	token        string
	refreshToken string
	done         chan struct{}
}

type TokenManagerConfig struct {
	ObtainTokens            func(context.Context) (string, string, error)
	RefreshWithRefreshToken func(context.Context, string) (string, string, error)
	Interval                time.Duration
}

func NewTokenManager(ctx context.Context, cfg TokenManagerConfig) (*TokenManager, error) {
	if cfg.ObtainTokens == nil {
		return nil, ErrTokenManagerNoObtainTokens
	}
	if cfg.RefreshWithRefreshToken == nil {
		return nil, ErrTokenManagerNoRefreshWithRefresh
	}
	if cfg.Interval <= 0 {
		cfg.Interval = 4 * time.Minute
	}

	tm := &TokenManager{
		obtainTokens:          cfg.ObtainTokens,
		refreshWithRefreshTok: cfg.RefreshWithRefreshToken,
		interval:              cfg.Interval,
		done:                  make(chan struct{}),
	}

	if err := tm.refresh(ctx); err != nil {
		return nil, err
	}

	go tm.run(ctx)
	return tm, nil
}

func (tm *TokenManager) run(ctx context.Context) {
	ticker := time.NewTicker(tm.interval)
	defer ticker.Stop()
	defer close(tm.done)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = tm.refresh(ctx)
		}
	}
}

func (tm *TokenManager) GetToken() string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.token
}

func (tm *TokenManager) RefreshToken(ctx context.Context) error {
	return tm.refresh(ctx)
}

func (tm *TokenManager) refresh(ctx context.Context) error {
	tm.mu.RLock()
	refreshTok := tm.refreshToken
	tm.mu.RUnlock()

	if refreshTok != "" {
		access, newRefresh, err := tm.refreshWithRefreshTok(ctx, refreshTok)
		if err == nil {
			tm.mu.Lock()
			tm.token = access
			tm.refreshToken = newRefresh
			tm.mu.Unlock()
			return nil
		}
	}

	access, newRefresh, err := tm.obtainTokens(ctx)
	if err != nil {
		return err
	}
	tm.mu.Lock()
	tm.token = access
	tm.refreshToken = newRefresh
	tm.mu.Unlock()
	return nil
}

func (tm *TokenManager) Stop() {
	<-tm.done
}
