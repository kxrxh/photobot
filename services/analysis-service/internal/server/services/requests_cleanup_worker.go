package services

import (
	"context"
	"errors"
	"sync"

	"csort.ru/analysis-service/internal/requests"
)

const RequestsCleanupWorkerServiceName = "requests-cleanup"

type RequestsCleanupWorkerService struct {
	requests *requests.Service

	startOnce sync.Once
	bg        backgroundCancelLoop
}

func NewRequestsCleanupWorkerService(requestsSvc *requests.Service) *RequestsCleanupWorkerService {
	return &RequestsCleanupWorkerService{requests: requestsSvc}
}

func (s *RequestsCleanupWorkerService) Start(ctx context.Context) error {
	if s.requests == nil {
		return errors.New("requests service is nil")
	}
	s.startOnce.Do(func() {
		loopCtx, cancel := context.WithCancel(ctx)
		s.bg.bind(cancel)
		go s.requests.StartCleanup(loopCtx)
	})
	return nil
}

func (s *RequestsCleanupWorkerService) String() string { return RequestsCleanupWorkerServiceName }

func (s *RequestsCleanupWorkerService) State(context.Context) (string, error) {
	return s.bg.state()
}

func (s *RequestsCleanupWorkerService) Terminate(context.Context) error {
	s.bg.stop()
	return nil
}
