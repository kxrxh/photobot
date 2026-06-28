package testing

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"csort.ru/coffeebot/internal/minio"

	"github.com/rs/zerolog"
	"github.com/testcontainers/testcontainers-go"
	minioModule "github.com/testcontainers/testcontainers-go/modules/minio"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

type TB interface {
	Helper()
	Errorf(format string, args ...any)
	Fatalf(format string, args ...any)
}

const (
	postgresImage = "postgres:17"
	minioImage    = "minio/minio"

	pgDB        = "photobox_test"
	pgUser      = "user"
	pgPassword  = "password"
	minioUser   = "minioadmin"
	minioPass   = "minioadmin"
	minioBucket = "photobot"
)

func schemaPath() string {
	dir, err := os.Getwd()
	if err != nil {
		return "databases/schema.sql"
	}
	for {
		try := filepath.Join(dir, "databases", "schema.sql")
		if _, err := os.Stat(try); err == nil {
			return try
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "databases/schema.sql"
}

func startPostgres(ctx context.Context) (string, func(), error) {
	schema := schemaPath()

	ctr, err := postgres.Run(ctx, postgresImage,
		postgres.WithDatabase(pgDB),
		postgres.WithUsername(pgUser),
		postgres.WithPassword(pgPassword),
		postgres.WithInitScripts(schema),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return "", nil, fmt.Errorf("start postgres: %w", err)
	}

	connStr, err := ctr.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return "", nil, fmt.Errorf("postgres connection string: %w", err)
	}

	cleanup := func() {
		_ = testcontainers.TerminateContainer(ctr)
	}

	return connStr, cleanup, nil
}

func startMinIO(ctx context.Context) (*minio.Client, func(), error) {
	ctr, err := minioModule.Run(ctx, minioImage,
		minioModule.WithUsername(minioUser),
		minioModule.WithPassword(minioPass),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("start minio: %w", err)
	}

	endpoint, err := ctr.ConnectionString(ctx)
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return nil, nil, fmt.Errorf("minio connection string: %w", err)
	}

	host, portStr, err := net.SplitHostPort(endpoint)
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return nil, nil, fmt.Errorf("parse minio endpoint %q: %w", endpoint, err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return nil, nil, fmt.Errorf("parse minio port %q: %w", portStr, err)
	}

	client, err := minio.NewClient(
		ctx,
		minio.MinioClientConfig{
			Host:      host,
			Port:      port,
			AccessKey: minioUser,
			SecretKey: minioPass,
			Bucket:    minioBucket,
			UseSSL:    false,
		},
		zerolog.Nop(),
	)
	if err != nil {
		_ = testcontainers.TerminateContainer(ctr)
		return nil, nil, fmt.Errorf("minio client: %w", err)
	}

	cleanup := func() {
		_ = testcontainers.TerminateContainer(ctr)
	}

	return client, cleanup, nil
}
