package services

import (
	"context"
	"errors"

	"csort.ru/analysis-service/internal/storage"
)

const (
	TempMinIOServiceName     = "temp-minio"
	AnalysisMinIOServiceName = "analysis-minio"
)

type StorageService struct {
	name   string
	client *storage.Client
}

func NewStorageService(name string, client *storage.Client) *StorageService {
	return &StorageService{name: name, client: client}
}

func (s *StorageService) Start(context.Context) error {
	if s.client == nil {
		return errors.New("storage client is nil")
	}
	return nil
}

func (s *StorageService) String() string { return s.name }

func (s *StorageService) State(context.Context) (string, error) {
	if s.client == nil {
		return "unavailable", nil
	}
	return "ready", nil
}

func (s *StorageService) Terminate(context.Context) error { return nil }
