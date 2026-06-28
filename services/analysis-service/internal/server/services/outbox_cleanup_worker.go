package services

import (
	"context"
	"errors"
	"sync"

	"csort.ru/analysis-service/internal/requests"
)

const OutboxCleanupWorkerServiceName = "outbox-cleanup"

type OutboxCleanupWorkerService struct {
	outbox *requests.OutboxRelay

	startOnce sync.Once
	bg        backgroundCancelLoop
}

func NewOutboxCleanupWorkerService(outbox *requests.OutboxRelay) *OutboxCleanupWorkerService {
	return &OutboxCleanupWorkerService{outbox: outbox}
}

func (s *OutboxCleanupWorkerService) Start(ctx context.Context) error {
	if s.outbox == nil {
		return errors.New("outbox relay is nil")
	}
	s.startOnce.Do(func() {
		loopCtx, cancel := context.WithCancel(ctx)
		s.bg.bind(cancel)
		go s.outbox.StartCleanup(loopCtx)
	})
	return nil
}

func (s *OutboxCleanupWorkerService) String() string { return OutboxCleanupWorkerServiceName }

func (s *OutboxCleanupWorkerService) State(context.Context) (string, error) {
	return s.bg.state()
}

func (s *OutboxCleanupWorkerService) Terminate(context.Context) error {
	s.bg.stop()
	return nil
}
