package services

import (
	"context"
	"errors"
	"sync"

	"csort.ru/analysis-service/internal/database"
)

const RequestsDBServiceName = "requests-db"

type RequestsDatabaseService struct {
	db        *database.DB
	closeOnce sync.Once
}

func NewRequestsDatabaseService(db *database.DB) *RequestsDatabaseService {
	return &RequestsDatabaseService{db: db}
}

func (s *RequestsDatabaseService) Start(ctx context.Context) error {
	if s.db == nil {
		return errors.New("requests database is nil")
	}
	return s.db.Ping(ctx)
}

func (s *RequestsDatabaseService) String() string { return RequestsDBServiceName }

func (s *RequestsDatabaseService) State(ctx context.Context) (string, error) {
	if s.db == nil {
		return "unavailable", nil
	}
	if err := s.db.Ping(ctx); err != nil {
		return "disconnected", nil
	}
	return "connected", nil
}

func (s *RequestsDatabaseService) Terminate(context.Context) error {
	s.closeOnce.Do(func() {
		if s.db != nil {
			s.db.Close()
		}
	})
	return nil
}
