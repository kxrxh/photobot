package services

import (
	"context"
	"errors"
	"sync"
	"time"

	"csort.ru/analysis-service/internal/messaging"
)

const RabbitMQServiceName = "rabbitmq"

type RabbitMQService struct {
	client       *messaging.Client
	recoveryTick time.Duration
	closeOnce    sync.Once
}

func NewRabbitMQService(client *messaging.Client, recoveryTick time.Duration) *RabbitMQService {
	return &RabbitMQService{
		client:       client,
		recoveryTick: recoveryTick,
	}
}

func (s *RabbitMQService) Start(context.Context) error {
	if s.client == nil {
		return errors.New("rabbitmq client is nil")
	}
	if err := s.client.EnsureConnected(); err != nil {
		return err
	}
	s.client.StartAutoRecovery(s.recoveryTick)
	return nil
}

func (s *RabbitMQService) String() string { return RabbitMQServiceName }

func (s *RabbitMQService) State(context.Context) (string, error) {
	if s.client == nil {
		return "unavailable", nil
	}
	if !s.client.IsConnected() {
		return "disconnected", nil
	}
	return "connected", nil
}

func (s *RabbitMQService) Terminate(context.Context) error {
	var err error
	s.closeOnce.Do(func() {
		if s.client != nil {
			err = s.client.Close()
		}
	})
	return err
}
