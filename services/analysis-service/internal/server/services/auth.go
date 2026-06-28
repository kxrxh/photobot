package services

import (
	"context"
	"errors"
	"sync"

	"csort.ru/analysis-service/internal/api/auth"
)

const AuthClientServiceName = "auth-client"

type AuthClientService struct {
	client    *auth.Client
	startOnce sync.Once
	startErr  error
}

func NewAuthClientService(client *auth.Client) *AuthClientService {
	return &AuthClientService{client: client}
}

func (s *AuthClientService) Start(ctx context.Context) error {
	if s.client == nil {
		return errors.New("auth client is nil")
	}
	s.startOnce.Do(func() {
		s.startErr = s.client.Start(ctx)
	})
	return s.startErr
}

func (s *AuthClientService) String() string { return AuthClientServiceName }

func (s *AuthClientService) State(context.Context) (string, error) {
	if s.client == nil {
		return "unavailable", nil
	}
	if s.client.GetToken() == "" {
		return "starting", nil
	}
	return "ready", nil
}

func (s *AuthClientService) Terminate(context.Context) error {
	if s.client != nil {
		s.client.EndJWKSBackground()
	}
	return nil
}
