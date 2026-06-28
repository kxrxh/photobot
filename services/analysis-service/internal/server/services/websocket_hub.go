package services

import (
	"context"
	"errors"
	"sync"

	"csort.ru/analysis-service/internal/transport/ws"
)

const WebSocketHubServiceName = "websocket-hub"

type WebSocketHubService struct {
	hub          *ws.Hub
	shutdownOnce sync.Once
}

func NewWebSocketHubService(hub *ws.Hub) *WebSocketHubService {
	return &WebSocketHubService{hub: hub}
}

func (s *WebSocketHubService) Start(context.Context) error {
	if s.hub == nil {
		return errors.New("websocket hub is nil")
	}
	return nil
}

func (s *WebSocketHubService) String() string { return WebSocketHubServiceName }

func (s *WebSocketHubService) State(context.Context) (string, error) {
	if s.hub == nil {
		return "unavailable", nil
	}
	return "running", nil
}

func (s *WebSocketHubService) Terminate(context.Context) error {
	s.shutdownOnce.Do(func() {
		if s.hub != nil {
			s.hub.Shutdown()
		}
	})
	return nil
}
