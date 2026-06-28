package minio

import (
	"context"
	"io"
	"time"
)

type ObjectStorage interface {
	UploadObject(
		ctx context.Context,
		objectKey string,
		reader io.Reader,
		size int64,
		contentType string,
	) error
	DeleteObject(ctx context.Context, objectKey string) error
	GeneratePresignedURL(
		ctx context.Context,
		objectKey string,
		expiry time.Duration,
	) (string, error)
	// GeneratePresignedURLs signs many keys; result[i] corresponds to keys[i] (empty if keys[i] is empty).
	GeneratePresignedURLs(
		ctx context.Context,
		objectKeys []string,
		expiry time.Duration,
	) ([]string, error)
}

var _ ObjectStorage = (*Client)(nil)
