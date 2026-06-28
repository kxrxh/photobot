package services

import (
	"context"
	"sync"

	"csort.ru/analysis-service/internal/analysis"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/requests"
	"github.com/jackc/pgx/v5/pgxpool"
)

const StuckProcessingCleanupServiceName = "requests-stuck-processing-cleanup"

type StuckProcessingCleanupService struct {
	requests     *requests.Service
	requestsPool *pgxpool.Pool

	startOnce sync.Once
	bg        backgroundCancelLoop
}

func NewStuckProcessingCleanupService(
	requestsSvc *requests.Service,
	requestsPool *pgxpool.Pool,
) *StuckProcessingCleanupService {
	return &StuckProcessingCleanupService{
		requests:     requestsSvc,
		requestsPool: requestsPool,
	}
}

func (s *StuckProcessingCleanupService) Start(ctx context.Context) error {
	if s.requestsPool == nil || s.requests == nil {
		return nil
	}
	s.startOnce.Do(func() {
		loopCtx, cancel := context.WithCancel(ctx)
		s.bg.bind(cancel)
		cleanupLock := analysis.NewCleanupLock(
			s.requestsPool,
			logger.GetLogger("analysis.cleanup_lock"),
		)
		cleanupLock.StartPeriodicCleanup(loopCtx, s.requests.CleanupStuckProcessingRequests)
	})
	return nil
}

func (s *StuckProcessingCleanupService) String() string { return StuckProcessingCleanupServiceName }

func (s *StuckProcessingCleanupService) State(context.Context) (string, error) {
	if s.requestsPool == nil {
		return "disabled", nil
	}
	return s.bg.state()
}

func (s *StuckProcessingCleanupService) Terminate(context.Context) error {
	s.bg.stop()
	return nil
}
