package services

import (
	"context"
	"errors"
	"sync"

	"csort.ru/analysis-service/internal/requests"
)

const OutboxRelayWorkerServiceName = "outbox-relay"

type OutboxRelayWorkerService struct {
	outbox *requests.OutboxRelay

	startOnce sync.Once
	bg        backgroundCancelLoop
}

func NewOutboxRelayWorkerService(outbox *requests.OutboxRelay) *OutboxRelayWorkerService {
	return &OutboxRelayWorkerService{outbox: outbox}
}

func (s *OutboxRelayWorkerService) Start(ctx context.Context) error {
	if s.outbox == nil {
		return errors.New("outbox relay is nil")
	}
	s.startOnce.Do(func() {
		loopCtx, cancel := context.WithCancel(ctx)
		s.bg.bind(cancel)
		go s.outbox.Start(loopCtx)
	})
	return nil
}

func (s *OutboxRelayWorkerService) String() string { return OutboxRelayWorkerServiceName }

func (s *OutboxRelayWorkerService) State(context.Context) (string, error) {
	return s.bg.state()
}

func (s *OutboxRelayWorkerService) Terminate(context.Context) error {
	s.bg.stop()
	return nil
}
