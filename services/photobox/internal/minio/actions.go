package minio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/minio/minio-go/v7"
	"golang.org/x/sync/errgroup"
)

// GeneratePresignedURL returns a time-limited GET URL for objectKey.
func (c *Client) GeneratePresignedURL(
	ctx context.Context,
	objectKey string,
	expiry time.Duration,
) (string, error) {
	if c == nil || c.Client == nil {
		return "", errors.New("minio client is not initialized")
	}
	if objectKey == "" {
		return "", errors.New("object key cannot be empty")
	}
	if c.BucketName == "" {
		return "", errors.New("bucket name is not configured")
	}
	const maxExpiry = 7 * 24 * time.Hour
	if expiry <= 0 || expiry > maxExpiry {
		return "", fmt.Errorf("expiry must be > 0 and <= %s", maxExpiry)
	}

	reqParams := make(url.Values)

	ctxTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	internalPresignedURL, err := c.Client.PresignedGetObject(
		ctxTimeout,
		c.BucketName,
		objectKey,
		expiry,
		reqParams,
	)
	if err != nil {
		c.log.Error().
			Err(err).
			Str("objectKey", objectKey).
			Str("bucket", c.BucketName).
			Msg("Failed to generate internal presigned URL")
		return "", fmt.Errorf(
			"failed to generate internal presigned URL for object %s: %w",
			objectKey,
			err,
		)
	}

	if c.PublicEndpoint == nil {
		c.log.Debug().
			Str("objectKey", objectKey).
			Str("bucket", c.BucketName).
			Dur("expiry", expiry).
			Msg("Generated internal presigned URL (PublicEndpoint not configured)")
		return internalPresignedURL.String(), nil
	}

	publicPresignedURL := &url.URL{}
	*publicPresignedURL = *internalPresignedURL

	publicPresignedURL.Scheme = c.PublicEndpoint.Scheme
	publicPresignedURL.Host = c.PublicEndpoint.Host

	basePath := strings.TrimSuffix(c.PublicEndpoint.Path, "/")
	if basePath != "" && !strings.HasPrefix(basePath, "/") {
		basePath = "/" + basePath
	}
	// Use EscapedPath so encoded segments stay intact (path.Clean would change them).
	objectPath := internalPresignedURL.EscapedPath()
	publicPresignedURL.Path = basePath + objectPath

	c.log.Debug().
		Str("objectKey", objectKey).
		Str("bucket", c.BucketName).
		Dur("expiry", expiry).
		Str("publicURL", publicPresignedURL.String()).
		Msg("Generated public presigned URL")
	return publicPresignedURL.String(), nil
}

const presignParallelism = 8

// GeneratePresignedURLs signs keys in parallel; empty keys stay empty in the result slice.
func (c *Client) GeneratePresignedURLs(
	ctx context.Context,
	objectKeys []string,
	expiry time.Duration,
) ([]string, error) {
	if len(objectKeys) == 0 {
		return nil, nil
	}
	if c == nil || c.Client == nil {
		return nil, errors.New("minio client is not initialized")
	}
	out := make([]string, len(objectKeys))
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(presignParallelism)
	for i, key := range objectKeys {
		i, key := i, key
		if key == "" {
			continue
		}
		g.Go(func() error {
			u, err := c.GeneratePresignedURL(gctx, key, expiry)
			if err != nil {
				return err
			}
			out[i] = u
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}

// DeleteObject removes objectKey from the configured bucket.
func (c *Client) DeleteObject(ctx context.Context, objectKey string) error {
	if c == nil || c.Client == nil {
		return errors.New("minio client is not initialized")
	}
	if objectKey == "" {
		return errors.New("object key cannot be empty")
	}
	if c.BucketName == "" {
		return errors.New("bucket name is not configured")
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	err := c.Client.RemoveObject(ctxTimeout, c.BucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		c.log.Error().
			Err(err).
			Str("objectKey", objectKey).
			Str("bucket", c.BucketName).
			Msg("Failed to delete object from MinIO")
		return fmt.Errorf(
			"failed to delete object %s from bucket %s: %w",
			objectKey,
			c.BucketName,
			err,
		)
	}

	c.log.Debug().
		Str("objectKey", objectKey).
		Str("bucket", c.BucketName).
		Msg("Successfully deleted object from MinIO")
	return nil
}

// UploadObject writes reader to objectKey in the configured bucket.
func (c *Client) UploadObject(
	ctx context.Context,
	objectKey string,
	reader io.Reader,
	size int64,
	contentType string,
) error {
	if c == nil || c.Client == nil {
		return errors.New("minio client is not initialized")
	}
	if objectKey == "" {
		return errors.New("object key cannot be empty")
	}
	if reader == nil {
		return errors.New("reader cannot be nil")
	}
	if size < -1 {
		return errors.New("size cannot be less than -1")
	}
	if c.BucketName == "" {
		return errors.New("bucket name is not configured")
	}

	if contentType == "" {
		if rs, ok := reader.(io.ReadSeeker); ok {
			if curr, err := rs.Seek(0, io.SeekCurrent); err == nil {
				buf := make([]byte, 3072)
				n, _ := rs.Read(buf)
				_, _ = rs.Seek(curr, io.SeekStart)
				if n > 0 {
					if mt := mimetype.Detect(buf[:n]); mt != nil {
						contentType = mt.String()
					}
				}
			}
		}
		if contentType == "" {
			contentType = "application/octet-stream"
		}
	}

	ctxTimeout, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	_, err := c.Client.PutObject(
		ctxTimeout,
		c.BucketName,
		objectKey,
		reader,
		size,
		minio.PutObjectOptions{ContentType: contentType},
	)
	if err != nil {
		c.log.Error().
			Err(err).
			Str("objectKey", objectKey).
			Str("bucket", c.BucketName).
			Int64("size", size).
			Str("contentType", contentType).
			Msg("Failed to upload object to MinIO")
		return fmt.Errorf(
			"failed to upload object %s to bucket %s: %w",
			objectKey,
			c.BucketName,
			err,
		)
	}

	c.log.Debug().
		Str("objectKey", objectKey).
		Str("bucket", c.BucketName).
		Int64("size", size).
		Str("contentType", contentType).
		Msg("Successfully uploaded object to MinIO")
	return nil
}
