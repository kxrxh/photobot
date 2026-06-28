package storage

import "github.com/minio/minio-go/v7"

type MinIOConfig struct {
	Host         string `validate:"required"`
	Port         int    `validate:"required,gte=1,lte=65535"`
	RootUser     string `validate:"required"`
	RootPassword string `validate:"required"`
	Bucket       string `validate:"required"`
	UseSSL       bool   `validate:"required"`
}

type ImageStorageConfig struct {
	Host         string `validate:"required"`
	Port         int    `validate:"required,gte=1,lte=65535"`
	RootUser     string `validate:"required"`
	RootPassword string `validate:"required"`
	Bucket       string `validate:"required"`
	UseSSL       bool   `validate:"required"`
	ExternalHost string
}

type Client struct {
	client       *minio.Client
	bucket       string
	externalHost string
}

type UploadResult struct {
	FileID   string            `json:"file_id"`
	FileName string            `json:"file_name"`
	URL      string            `json:"url"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type TempObjectRef struct {
	ID           string
	PresignedURL string
}
