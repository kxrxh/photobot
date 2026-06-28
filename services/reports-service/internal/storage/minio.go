package storage

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"csort.ru/reports-service/internal/config"
	"csort.ru/reports-service/internal/domain"
	"csort.ru/reports-service/internal/objectstore"

	"github.com/minio/minio-go/v7"
)

type MinIOService struct {
	store  *objectstore.Client
	public *PresignedPublicBase
}

func NewMinIOService(cfg config.MinIOConfig) (*MinIOService, error) {
	pub, err := ParsePresignedPublicBase(cfg.PublicBaseURL)
	if err != nil {
		return nil, fmt.Errorf("minio public url: %w", err)
	}
	c, err := objectstore.New(cfg)
	if err != nil {
		return nil, err
	}
	return &MinIOService{store: c, public: pub}, nil
}

func (s *MinIOService) Ping(ctx context.Context) error {
	return s.store.Ping(ctx)
}

func (s *MinIOService) PresignedReportDownloadURL(
	ctx context.Context,
	analysisID, format string,
) (string, time.Duration, error) {
	key, err := ReportObjectKey(analysisID, format)
	if err != nil {
		return "", 0, err
	}
	ct, err := reportContentType(format)
	if err != nil {
		return "", 0, err
	}
	fn, err := ReportAttachmentFilename(analysisID, format)
	if err != nil {
		return "", 0, err
	}
	exp := presignTTLDuration()
	params := url.Values{}
	params.Set("response-content-type", ct)
	params.Set("response-content-disposition", `attachment; filename="`+fn+`"`)
	u, err := s.store.PresignedGetObject(ctx, key, exp, params)
	if err != nil {
		return "", 0, err
	}
	out, err := RewritePresignedToPublic(u.String(), s.public)
	if err != nil {
		return "", 0, err
	}
	return out, exp, nil
}

func (s *MinIOService) UploadBuffer(
	ctx context.Context,
	fileName string,
	body []byte,
	contentType string,
	metadata map[string]string,
) error {
	return s.store.Put(ctx, fileName, body, contentType, metadata)
}

func (s *MinIOService) GetFileStatus(
	ctx context.Context,
	fileName string,
) (domain.MinIOFileStatus, error) {
	meta, err := s.store.Stat(ctx, fileName)
	if err != nil {
		if errors.Is(err, objectstore.ErrNotFound) {
			return domain.MinIOFileStatus{Exists: false}, nil
		}
		return domain.MinIOFileStatus{}, err
	}
	out := map[string]string{}
	for k, v := range meta.UserMetadata {
		out[strings.ToLower(k)] = v
	}
	return domain.MinIOFileStatus{
		Exists:   true,
		Metadata: out,
	}, nil
}

func (s *MinIOService) GetObjectReader(
	ctx context.Context,
	objectKey string,
) (*minio.Object, error) {
	return s.store.GetObject(ctx, objectKey)
}
