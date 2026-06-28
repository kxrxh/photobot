package objectstore

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"csort.ru/reports-service/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var ErrNotFound = errors.New("objectstore: object not found")

type ObjectMeta struct {
	UserMetadata map[string]string
}

type Client struct {
	client     *minio.Client
	bucket     string
	ensureOnce sync.Once
	ensureErr  error
}

func New(cfg config.MinIOConfig) (*Client, error) {
	endpoint := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.RootUser, cfg.RootPassword, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Client{client: mc, bucket: cfg.Bucket}, nil
}

func (c *Client) Ping(ctx context.Context) error {
	_, err := c.client.BucketExists(ctx, c.bucket)
	return err
}

func (c *Client) ensureBucket(ctx context.Context) error {
	c.ensureOnce.Do(func() {
		exists, err := c.client.BucketExists(ctx, c.bucket)
		if err != nil {
			c.ensureErr = err
			return
		}
		if exists {
			return
		}
		c.ensureErr = c.client.MakeBucket(
			ctx,
			c.bucket,
			minio.MakeBucketOptions{Region: "us-east-1"},
		)
	})
	return c.ensureErr
}

func (c *Client) Put(
	ctx context.Context,
	objectKey string,
	body []byte,
	contentType string,
	userMetadata map[string]string,
) error {
	if err := c.ensureBucket(ctx); err != nil {
		return err
	}
	meta := map[string]string{}
	for k, v := range userMetadata {
		meta[k] = v
	}
	_, err := c.client.PutObject(
		ctx,
		c.bucket,
		objectKey,
		bytes.NewReader(body),
		int64(len(body)),
		minio.PutObjectOptions{
			ContentType:  contentType,
			UserMetadata: meta,
		},
	)
	return err
}

func (c *Client) GetObject(ctx context.Context, objectKey string) (*minio.Object, error) {
	if err := c.ensureBucket(ctx); err != nil {
		return nil, err
	}
	o, err := c.client.GetObject(ctx, c.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, mapObjectErr(err, objectKey)
	}
	return o, nil
}

func (c *Client) Stat(ctx context.Context, objectKey string) (ObjectMeta, error) {
	stat, err := c.client.StatObject(ctx, c.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return ObjectMeta{}, mapObjectErr(err, objectKey)
	}
	meta := map[string]string{}
	for k, v := range stat.UserMetadata {
		meta[k] = v
	}
	return ObjectMeta{UserMetadata: meta}, nil
}

func mapObjectErr(err error, objectKey string) error {
	if err == nil {
		return nil
	}
	resp := minio.ToErrorResponse(err)
	if resp.Code == "NoSuchKey" || resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%w: %s", ErrNotFound, objectKey)
	}
	return err
}
