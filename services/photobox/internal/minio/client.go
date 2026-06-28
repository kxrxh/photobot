package minio

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog"
)

const (
	minioConnectionTimeout = 5 * time.Second
)

type MinioClientConfig struct {
	Host           string
	Port           int
	AccessKey      string //nolint:gosec // G117: config struct, creds from env
	SecretKey      string //nolint:gosec // G117: config struct, creds from env
	Bucket         string
	UseSSL         bool
	PublicEndpoint string
}

type Client struct {
	Client         *minio.Client
	BucketName     string
	PublicEndpoint *url.URL
	log            zerolog.Logger
}

func configIsMissing(cfg MinioClientConfig) bool {
	return cfg.Host == "" || cfg.Port == 0 || cfg.AccessKey == "" || cfg.SecretKey == "" ||
		cfg.Bucket == ""
}

func NewClient(ctx context.Context, cfg MinioClientConfig, zlog zerolog.Logger) (*Client, error) {
	if configIsMissing(cfg) {
		zlog.Warn().
			Msg("Essential MinIO configuration fields (MINIO_HOST, MINIO_PORT, MINIO_ROOT_USER, MINIO_ROOT_PASSWORD, MINIO_BUCKET) are not fully set. MinIO functionality will be disabled.")
		return nil, nil
	}

	minioEndpoint := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	minioClient, err := minio.New(minioEndpoint, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:       cfg.UseSSL,
		BucketLookup: minio.BucketLookupPath,
	})
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to initialize MinIO client")
		return nil, fmt.Errorf("failed to initialize MinIO client: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, minioConnectionTimeout)
	defer cancel()
	_, err = minioClient.ListBuckets(ctx)
	if err != nil {
		zlog.Error().
			Err(err).
			Msg("Failed to connect to MinIO or list buckets. Check endpoint and credentials.")
		return nil, fmt.Errorf("failed to connect to MinIO: %w", err)
	}

	var parsedPublicEndpoint *url.URL
	if cfg.PublicEndpoint != "" {
		endpointToParse := cfg.PublicEndpoint
		if !strings.HasPrefix(endpointToParse, "http") {
			scheme := "https://"
			if !cfg.UseSSL {
				scheme = "http://"
			}
			endpointToParse = scheme + endpointToParse
			zlog.Info().
				Str("originalEndpoint", cfg.PublicEndpoint).
				Str("parsedAs", endpointToParse).
				Msg("Prepended scheme to MINIO_PUBLIC_ENDPOINT")
		}

		parsedPublicEndpoint, err = url.Parse(endpointToParse)
		if err != nil {
			zlog.Error().
				Err(err).
				Str("publicEndpoint", cfg.PublicEndpoint).
				Msg("Failed to parse PublicEndpoint URL")
			parsedPublicEndpoint = nil
		}
	}

	clientWrapper := &Client{
		Client:         minioClient,
		BucketName:     cfg.Bucket,
		PublicEndpoint: parsedPublicEndpoint,
		log:            zlog.With().Str("component", "minio").Logger(),
	}

	err = clientWrapper.ensureBucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure bucket exists: %w", err)
	}

	return clientWrapper, nil
}

func (c *Client) GetBucketName() string {
	if c == nil {
		return ""
	}
	return c.BucketName
}

func (c *Client) GetClient() *minio.Client {
	if c == nil {
		return nil
	}
	return c.Client
}

func (c *Client) ensureBucketExists(ctx context.Context, bucketName string) error {
	if c.Client == nil {
		return errors.New("minio client not initialized")
	}

	exists, err := c.Client.BucketExists(ctx, bucketName)
	if err != nil {
		c.log.Error().
			Err(err).
			Str("bucketName", bucketName).
			Msg("Failed to check if bucket exists")
		return err
	}

	if exists {
		c.log.Debug().Str("bucketName", bucketName).Msg("Bucket already exists.")
		return nil
	}

	err = c.Client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{
		ObjectLocking: false,
	})
	if err != nil {
		c.log.Error().Err(err).Str("bucketName", bucketName).Msg("Failed to create bucket")
		return err
	}

	c.log.Info().Str("bucketName", bucketName).Msg("Successfully created bucket")
	return nil
}
