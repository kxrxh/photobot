package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"csort.ru/analysis-service/internal/fileutil"
	"csort.ru/analysis-service/internal/imageutil"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func minioHTTPTransport() http.RoundTripper {
	t, ok := http.DefaultTransport.(*http.Transport)
	if !ok {
		return http.DefaultTransport
	}
	tr := t.Clone()
	tr.MaxIdleConns = 128
	tr.MaxIdleConnsPerHost = 64
	return tr
}

func New(ctx context.Context, cfg *ImageStorageConfig) (*Client, error) {
	client, err := minio.New(
		fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		&minio.Options{
			Creds:     credentials.NewStaticV4(cfg.RootUser, cfg.RootPassword, ""),
			Secure:    cfg.UseSSL,
			Transport: minioHTTPTransport(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create MinIO client for host %s:%d: %w",
			cfg.Host,
			cfg.Port,
			err,
		)
	}

	_, err = client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MinIO at %s:%d: %w", cfg.Host, cfg.Port, err)
	}

	err = client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{})
	if err != nil {
		exists, errBucketExists := client.BucketExists(ctx, cfg.Bucket)
		if errBucketExists != nil || !exists {
			return nil, fmt.Errorf("failed to create bucket %s: %w", cfg.Bucket, err)
		}
	}

	return &Client{
		client:       client,
		bucket:       cfg.Bucket,
		externalHost: cfg.ExternalHost,
	}, nil
}

func NewForRead(ctx context.Context, cfg *ImageStorageConfig) (*Client, error) {
	client, err := minio.New(
		fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		&minio.Options{
			Creds:     credentials.NewStaticV4(cfg.RootUser, cfg.RootPassword, ""),
			Secure:    cfg.UseSSL,
			Transport: minioHTTPTransport(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to create MinIO client for host %s:%d: %w",
			cfg.Host,
			cfg.Port,
			err,
		)
	}

	_, err = client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MinIO at %s:%d: %w", cfg.Host, cfg.Port, err)
	}

	return &Client{
		client:       client,
		bucket:       cfg.Bucket,
		externalHost: cfg.ExternalHost,
	}, nil
}

func (c *Client) GetObject(ctx context.Context, key string) ([]byte, string, error) {
	obj, err := c.client.GetObject(ctx, c.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object %s from bucket %s: %w", key, c.bucket, err)
	}
	defer func() { _ = obj.Close() }()

	info, err := obj.Stat()
	if err != nil {
		return nil, "", fmt.Errorf("failed to stat object %s: %w", key, err)
	}

	data, err := io.ReadAll(obj)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read object %s: %w", key, err)
	}

	contentType := info.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return data, contentType, nil
}

func (c *Client) ObjectExists(ctx context.Context, key string) (bool, error) {
	_, err := c.client.StatObject(ctx, c.bucket, key, minio.StatObjectOptions{})
	if err == nil {
		return true, nil
	}
	resp := minio.ToErrorResponse(err)
	if resp.Code == "NoSuchKey" || resp.StatusCode == http.StatusNotFound {
		return false, nil
	}
	return false, fmt.Errorf("failed to stat object %s: %w", key, err)
}

func (c *Client) GetObjectStream(ctx context.Context, key string) (io.ReadCloser, string, error) {
	obj, err := c.client.GetObject(ctx, c.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get object %s from bucket %s: %w", key, c.bucket, err)
	}

	info, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, "", fmt.Errorf("failed to stat object %s: %w", key, err)
	}

	contentType := info.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	if contentType == "application/octet-stream" {
		peek := make([]byte, 12)
		n, readErr := io.ReadFull(obj, peek)
		if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			_ = obj.Close()
			return nil, "", fmt.Errorf("failed to read object %s: %w", key, readErr)
		}
		contentType = imageutil.GetMimeTypeFromBytes(peek[:n])
		return &streamReadCloser{
			Reader: io.MultiReader(bytes.NewReader(peek[:n]), obj),
			close:  obj.Close,
		}, contentType, nil
	}

	return obj, contentType, nil
}

func rewritePresignedURLForPublicAccess(internal, publicBase string) (string, error) {
	publicBase = strings.TrimSpace(publicBase)
	if publicBase == "" {
		return internal, nil
	}
	baseURL, err := url.Parse(strings.TrimSuffix(publicBase, "/"))
	if err != nil {
		return "", fmt.Errorf("MINIO_IMAGES_EXTERNAL_HOST: %w", err)
	}
	if baseURL.Scheme == "" || baseURL.Host == "" {
		return "", errors.New("MINIO_IMAGES_EXTERNAL_HOST must include scheme and host")
	}
	uInternal, err := url.Parse(internal)
	if err != nil {
		return "", fmt.Errorf("parse presigned url: %w", err)
	}
	prefix := strings.TrimSuffix(baseURL.Path, "/")
	if prefix == "" || prefix == "/" {
		prefix = ""
	}
	ip := uInternal.Path
	if ip == "" {
		ip = "/"
	}
	if !strings.HasPrefix(ip, "/") {
		ip = "/" + ip
	}
	out := url.URL{
		Scheme:   baseURL.Scheme,
		Host:     baseURL.Host,
		Path:     prefix + ip,
		RawQuery: uInternal.RawQuery,
	}
	return out.String(), nil
}

func (c *Client) PresignedGetObject(
	ctx context.Context,
	key string,
	expiry time.Duration,
) (string, error) {
	if key == "" {
		return "", errors.New("object key cannot be empty")
	}
	u, err := c.client.PresignedGetObject(ctx, c.bucket, key, expiry, url.Values{})
	if err != nil {
		return "", fmt.Errorf("failed to presign object %s: %w", key, err)
	}
	if c.externalHost != "" {
		return rewritePresignedURLForPublicAccess(u.String(), c.externalHost)
	}
	return u.String(), nil
}

const (
	objectPrefix = "object_"
	objectSuffix = ".jpg"
)

func parseObjectID(key string) (string, bool) {
	base := key
	if i := strings.LastIndex(key, "/"); i >= 0 {
		base = key[i+1:]
	}
	if !strings.HasPrefix(base, objectPrefix) || !strings.HasSuffix(base, objectSuffix) {
		return "", false
	}
	id := strings.TrimSuffix(strings.TrimPrefix(base, objectPrefix), objectSuffix)
	if id == "" {
		return "", false
	}
	return id, true
}

func (c *Client) ListTempAnalysisObjects(
	ctx context.Context,
	tempID string,
	expiry time.Duration,
) ([]TempObjectRef, error) {
	prefix := tempID + "/objects/"
	opts := minio.ListObjectsOptions{Prefix: prefix, Recursive: true}
	matching := make([]string, 0, 32)
	for info := range c.client.ListObjects(ctx, c.bucket, opts) {
		if info.Err != nil {
			return nil, fmt.Errorf("list objects for prefix %s: %w", prefix, info.Err)
		}
		if _, ok := parseObjectID(info.Key); ok {
			matching = append(matching, info.Key)
		}
	}
	result := make([]TempObjectRef, len(matching))
	for i, key := range matching {
		id, _ := parseObjectID(key)
		presignedURL, err := c.PresignedGetObject(ctx, key, expiry)
		if err != nil {
			return nil, fmt.Errorf("presign %s: %w", key, err)
		}
		result[i] = TempObjectRef{ID: id, PresignedURL: presignedURL}
	}
	sort.Slice(result, func(i, j int) bool {
		ni, erri := strconv.ParseInt(result[i].ID, 10, 64)
		nj, errj := strconv.ParseInt(result[j].ID, 10, 64)
		if erri == nil && errj == nil {
			return ni < nj
		}
		return result[i].ID < result[j].ID
	})
	return result, nil
}

type streamReadCloser struct {
	io.Reader
	close func() error
}

func (s *streamReadCloser) Close() error { return s.close() }

func (c *Client) UploadMultipartFile(
	ctx context.Context,
	fileHeader *multipart.FileHeader,
) (*UploadResult, error) {
	if fileHeader == nil {
		return nil, errors.New("file header is nil")
	}
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", fileHeader.Filename, err)
	}
	defer func() { _ = file.Close() }()

	prefix := make([]byte, fileutil.MaxPrefixBytes)
	n, readErr := io.ReadFull(file, prefix)
	prefix = prefix[:n]
	if n == 0 {
		return nil, fmt.Errorf("file %s is empty", fileHeader.Filename)
	}
	if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
		return nil, fmt.Errorf("failed to read file %s: %w", fileHeader.Filename, readErr)
	}

	contentType, err := fileutil.ValidateFileContentFromPrefix(prefix, fileHeader.Filename)
	if err != nil {
		return nil, err
	}

	body := io.MultiReader(bytes.NewReader(prefix), file)
	return c.UploadFile(ctx, body, fileHeader.Filename, contentType, nil)
}

func (c *Client) UploadFile(
	ctx context.Context,
	reader io.Reader,
	filename string,
	contentType string,
	metadata map[string]string,
) (*UploadResult, error) {
	fileID := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(filename))

	options := minio.PutObjectOptions{
		ContentType:  contentType,
		UserMetadata: metadata,
	}

	_, err := c.client.PutObject(ctx, c.bucket, fileID, reader, -1, options)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to MinIO: %w", err)
	}

	url := c.bucket + "/" + fileID

	return &UploadResult{
		FileID:   fileID,
		FileName: filename,
		URL:      url,
		Metadata: metadata,
	}, nil
}

func (c *Client) DeleteFile(ctx context.Context, fileID string) error {
	err := c.client.RemoveObject(ctx, c.bucket, fileID, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file from MinIO: %w", err)
	}

	return nil
}

func (c *Client) DeleteFiles(ctx context.Context, fileIDs []string) {
	if len(fileIDs) == 0 {
		return
	}
	objectsCh := make(chan minio.ObjectInfo, len(fileIDs))
	go func() {
		defer close(objectsCh)
		for _, fileID := range fileIDs {
			objectsCh <- minio.ObjectInfo{Key: fileID}
		}
	}()
	for err := range c.client.RemoveObjects(ctx, c.bucket, objectsCh, minio.RemoveObjectsOptions{}) {
		_ = err.Err
	}
}
