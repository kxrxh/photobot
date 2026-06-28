package requests

import (
	"context"
	"time"

	"csort.ru/analysis-service/internal/repository/requests"
)

const (
	finalizeQueueSize   = 64
	finalizeMaxAttempts = 20
)

type finalizeTask struct {
	ctx    context.Context
	params requests.MarkRequestAsCompletedIfWaitingParams
}

func (s *Service) ensureFinalizeWorker(ctx context.Context) {
	s.finalizeOnce.Do(func() {
		s.finalizeCh = make(chan finalizeTask, finalizeQueueSize)
		go s.runFinalizeWorker(ctx)
	})
}

func (s *Service) runFinalizeWorker(fallback context.Context) {
	for task := range s.finalizeCh {
		if task.ctx != nil {
			//nolint:contextcheck // ctx is captured at enqueue from the confirm handler.
			s.finalizeOne(task.ctx, task.params)
			continue
		}
		s.finalizeOne(fallback, task.params)
	}
}

func (s *Service) finalizeOne(
	ctx context.Context,
	params requests.MarkRequestAsCompletedIfWaitingParams,
) {
	if ctx == nil {
		return
	}
	bgCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	for attempt := 1; attempt <= finalizeMaxAttempts; attempt++ {
		_, err := s.repo.MarkRequestAsCompletedIfWaiting(bgCtx, params)
		if err == nil {
			return
		}
		time.Sleep(time.Duration(attempt*250) * time.Millisecond)
	}
}

func (s *Service) enqueueFinalize(ctx context.Context, task finalizeTask) {
	task.ctx = ctx
	s.ensureFinalizeWorker(ctx)
	select {
	case s.finalizeCh <- task:
	default:
		go func() { s.finalizeCh <- task }()
	}
}
