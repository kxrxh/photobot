package image

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"csort.ru/analysis-service/internal/cache"
	"csort.ru/analysis-service/internal/imageutil"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/repository/kalibr"
	"csort.ru/analysis-service/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

var (
	ErrNotFound        = errors.New("not found")
	ErrIndexOutOfRange = errors.New("index out of range")
)

const analysisCacheTTL = 60 * time.Second

type analysisBucketClient interface {
	GetObjectStream(ctx context.Context, key string) (io.ReadCloser, string, error)
	ObjectExists(ctx context.Context, key string) (bool, error)
	PresignedGetObject(ctx context.Context, key string, expiry time.Duration) (string, error)
}

type Service struct {
	kalibrQueries         *kalibr.Queries
	analysisStorageClient analysisBucketClient
	analysisCache         *cache.TTLCache[string, *analysisCacheEntry]
	logger                zerolog.Logger
}

func NewImageService(
	kalibrQueries *kalibr.Queries,
	analysisStorageClient *storage.Client,
) *Service {
	return &Service{
		kalibrQueries:         kalibrQueries,
		analysisStorageClient: analysisStorageClient,
		analysisCache:         cache.NewTTLCache[string, *analysisCacheEntry](analysisCacheTTL),
		logger:                logger.GetLogger("image.service"),
	}
}

func (s *Service) GetSourceStream(
	ctx context.Context,
	analysisID string,
	index int,
) (io.ReadCloser, string, error) {
	if err := s.validateInput(analysisID, index); err != nil {
		return nil, "", err
	}
	entry, err := s.getAnalysisFilesOrError(ctx, analysisID)
	if err != nil {
		return nil, "", err
	}
	key, err := s.resolveImageKeyWithFallback(
		ctx,
		analysisID,
		index,
		entry.FilesSource,
		imageutil.SourceKey,
	)
	if err != nil {
		return nil, "", err
	}
	return s.analysisStorageClient.GetObjectStream(ctx, key)
}

func (s *Service) GetOutputStream(
	ctx context.Context,
	analysisID string,
	index int,
) (io.ReadCloser, string, error) {
	if err := s.validateInput(analysisID, index); err != nil {
		return nil, "", err
	}
	entry, err := s.getAnalysisFilesOrError(ctx, analysisID)
	if err != nil {
		return nil, "", err
	}
	key, err := s.resolveImageKeyWithFallback(
		ctx,
		analysisID,
		index,
		entry.FilesOutput,
		imageutil.OutputKey,
	)
	if err != nil {
		return nil, "", err
	}
	return s.analysisStorageClient.GetObjectStream(ctx, key)
}

func (s *Service) GetSourcePresignedURL(
	ctx context.Context,
	analysisID string,
	index int,
) (string, error) {
	if err := s.validateInput(analysisID, index); err != nil {
		return "", err
	}
	entry, err := s.getAnalysisFilesOrError(ctx, analysisID)
	if err != nil {
		return "", err
	}
	return s.presignSourceAtIndex(ctx, analysisID, index, entry.FilesSource)
}

func (s *Service) GetSourcePresignedURLWithFiles(
	ctx context.Context,
	analysisID string,
	index int,
	filesSource []string,
) (string, error) {
	if err := s.validateInput(analysisID, index); err != nil {
		return "", err
	}
	return s.presignSourceAtIndex(ctx, analysisID, index, filesSource)
}

func (s *Service) presignSourceAtIndex(
	ctx context.Context,
	analysisID string,
	index int,
	filesSource []string,
) (string, error) {
	key, err := s.resolveImageKeyWithFallback(
		ctx,
		analysisID,
		index,
		filesSource,
		imageutil.SourceKey,
	)
	if err != nil {
		return "", err
	}
	return s.analysisStorageClient.PresignedGetObject(ctx, key, time.Hour)
}

func (s *Service) GetOutputPresignedURL(
	ctx context.Context,
	analysisID string,
	index int,
) (string, error) {
	if err := s.validateInput(analysisID, index); err != nil {
		return "", err
	}
	entry, err := s.getAnalysisFilesOrError(ctx, analysisID)
	if err != nil {
		return "", err
	}
	key, err := s.resolveImageKeyWithFallback(
		ctx,
		analysisID,
		index,
		entry.FilesOutput,
		imageutil.OutputKey,
	)
	if err != nil {
		return "", err
	}
	return s.analysisStorageClient.PresignedGetObject(ctx, key, time.Hour)
}

func (s *Service) GetObjectPresignedURL(
	ctx context.Context,
	analysisID string,
	objectFile string,
) (string, error) {
	if analysisID == "" {
		return "", errors.New("analysisID cannot be empty")
	}
	if objectFile == "" {
		return "", ErrNotFound
	}
	key := imageutil.ObjectKey(analysisID, objectFile)
	return s.analysisStorageClient.PresignedGetObject(ctx, key, time.Hour)
}

func (s *Service) GetObjectStream(
	ctx context.Context,
	analysisID string,
	objectID int32,
) (io.ReadCloser, string, error) {
	if analysisID == "" {
		return nil, "", errors.New("analysisID cannot be empty")
	}
	entry, err := s.getAnalysisFilesOrError(ctx, analysisID)
	if err != nil {
		return nil, "", err
	}
	filename, err := s.objectFileAt(entry.ObjectsJSON, int(objectID))
	if err != nil {
		return nil, "", err
	}
	key := imageutil.ObjectKey(analysisID, filename)
	return s.analysisStorageClient.GetObjectStream(ctx, key)
}

func (s *Service) GetObjectStreamByFile(
	ctx context.Context,
	analysisID string,
	objectFile string,
) (io.ReadCloser, string, error) {
	if analysisID == "" {
		return nil, "", errors.New("analysisID cannot be empty")
	}
	filename := strings.TrimSpace(objectFile)
	if filename == "" {
		return nil, "", ErrNotFound
	}
	key := imageutil.ObjectKey(analysisID, filename)
	return s.analysisStorageClient.GetObjectStream(ctx, key)
}

func (s *Service) objectFileAt(objectsJSON []byte, index int) (string, error) {
	var objs []struct {
		File string `json:"file"`
	}
	if err := json.Unmarshal(objectsJSON, &objs); err != nil {
		return "", fmt.Errorf("failed to unmarshal objects: %w", err)
	}
	if index < 0 || index >= len(objs) {
		return "", ErrIndexOutOfRange
	}
	f := objs[index].File
	if f == "" {
		return "", ErrNotFound
	}
	return f, nil
}

func (s *Service) validateInput(analysisID string, index int) error {
	if analysisID == "" {
		return errors.New("analysisID cannot be empty")
	}
	if index < 0 {
		return ErrIndexOutOfRange
	}
	return nil
}

func fallbackImageFilenames(index int) []string {
	return []string{fmt.Sprintf("image_%d.jpg", index)}
}

func (s *Service) resolveImageKeyWithFallback(
	ctx context.Context,
	analysisID string,
	index int,
	files []string,
	keyBuilder func(string, string) string,
) (string, error) {
	if index >= 0 && index < len(files) {
		filename := strings.TrimSpace(files[index])
		if filename != "" {
			return keyBuilder(analysisID, filename), nil
		}
	}

	for _, fallback := range fallbackImageFilenames(index) {
		key := keyBuilder(analysisID, fallback)
		exists, err := s.analysisStorageClient.ObjectExists(ctx, key)
		if err != nil {
			continue
		}
		if exists {
			return key, nil
		}
	}

	if index >= len(files) {
		return "", ErrIndexOutOfRange
	}
	return "", ErrNotFound
}

func (s *Service) getAnalysisFilesOrError(
	ctx context.Context,
	analysisID string,
) (*analysisCacheEntry, error) {
	if cached, ok := s.analysisCache.Get(analysisID); ok {
		return cached, nil
	}
	id, err := uuid.Parse(analysisID)
	if err != nil {
		s.logger.Warn().Str("analysisID", analysisID).Msg("Invalid analysis ID format")
		return nil, ErrNotFound
	}
	meta, err := s.kalibrQueries.GetAnalysisImageMetaByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn().Str("analysisID", analysisID).Msg("Analysis not found")
			return nil, ErrNotFound
		}
		s.logger.Error().Err(err).Str("analysisID", analysisID).Msg("Failed to get analysis")
		return nil, fmt.Errorf("failed to get analysis: %w", err)
	}
	entry := &analysisCacheEntry{
		FilesSource: meta.FilesSource,
		FilesOutput: meta.FilesOutput,
		ObjectsJSON: meta.Objects,
	}
	s.analysisCache.Set(analysisID, entry)
	return entry, nil
}
