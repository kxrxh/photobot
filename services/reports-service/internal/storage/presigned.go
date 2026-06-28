package storage

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"
)

type PresignedPublicBase struct {
	Scheme string
	Host   string
	Path   string
}

func ParsePresignedPublicBase(raw string) (*PresignedPublicBase, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil //nolint:nilnil // optional config
	}
	u, err := url.Parse(strings.TrimSuffix(raw, "/"))
	if err != nil {
		return nil, fmt.Errorf("MINIO_PUBLIC_BASE_URL: %w", err)
	}
	if u.Scheme == "" || u.Host == "" {
		return nil, errors.New("MINIO_PUBLIC_BASE_URL must include scheme and host")
	}
	path := u.Path
	if path == "" {
		path = "/"
	}
	return &PresignedPublicBase{
		Scheme: u.Scheme,
		Host:   u.Host,
		Path:   strings.TrimSuffix(path, "/"),
	}, nil
}

func RewritePresignedToPublic(internal string, pub *PresignedPublicBase) (string, error) {
	if pub == nil {
		return internal, nil
	}
	if pub.Scheme == "" || pub.Host == "" {
		return "", errors.New("presigned public base must include scheme and host")
	}
	uInternal, err := url.Parse(internal)
	if err != nil {
		return "", fmt.Errorf("parse presigned url: %w", err)
	}
	prefix := strings.TrimSuffix(pub.Path, "/")
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
		Scheme:   pub.Scheme,
		Host:     pub.Host,
		Path:     prefix + ip,
		RawQuery: uInternal.RawQuery,
	}
	return out.String(), nil
}

const presignTTL = 5 * time.Minute

func presignTTLDuration() time.Duration { return presignTTL }
