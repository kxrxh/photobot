package services

import (
	"context"
	"errors"
	"sync"

	"csort.ru/analysis-service/internal/database"
)

const KalibrDBServiceName = "kalibr-db"

type KalibrDatabaseService struct {
	db        *database.DB
	closeOnce sync.Once
}

func NewKalibrDatabaseService(db *database.DB) *KalibrDatabaseService {
	return &KalibrDatabaseService{db: db}
}

func (s *KalibrDatabaseService) Start(ctx context.Context) error {
	if s.db == nil {
		return errors.New("kalibr database is nil")
	}
	return s.db.Ping(ctx)
}

func (s *KalibrDatabaseService) String() string { return KalibrDBServiceName }

func (s *KalibrDatabaseService) State(ctx context.Context) (string, error) {
	if s.db == nil {
		return "unavailable", nil
	}
	if err := s.db.Ping(ctx); err != nil {
		return "disconnected", nil
	}
	return "connected", nil
}

func (s *KalibrDatabaseService) Terminate(context.Context) error {
	s.closeOnce.Do(func() {
		if s.db != nil {
			s.db.Close()
		}
	})
	return nil
}
