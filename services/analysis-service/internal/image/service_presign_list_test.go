package image

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"csort.ru/analysis-service/internal/cache"
	"csort.ru/analysis-service/internal/imageutil"
	"csort.ru/analysis-service/internal/logger"
)

type stubAnalysisBucket struct {
	lastPresignKey string
	streamOKKeys   map[string]struct{}
}

func (s *stubAnalysisBucket) GetObjectStream(
	ctx context.Context,
	key string,
) (io.ReadCloser, string, error) {
	if s.streamOKKeys == nil {
		return nil, "", errors.New("not found")
	}
	if _, ok := s.streamOKKeys[key]; ok {
		return io.NopCloser(strings.NewReader("x")), "image/jpeg", nil
	}
	return nil, "", errors.New("not found")
}

func (s *stubAnalysisBucket) ObjectExists(ctx context.Context, key string) (bool, error) {
	if s.streamOKKeys == nil {
		return false, nil
	}
	_, ok := s.streamOKKeys[key]
	return ok, nil
}

func (s *stubAnalysisBucket) PresignedGetObject(
	ctx context.Context,
	key string,
	expiry time.Duration,
) (string, error) {
	s.lastPresignKey = key
	return "https://signed.example/" + key, nil
}

func testImageServiceWithStub(stub *stubAnalysisBucket) *Service {
	return &Service{
		kalibrQueries:         nil,
		analysisStorageClient: stub,
		analysisCache:         cache.NewTTLCache[string, *analysisCacheEntry](time.Second),
		logger:                logger.GetLogger("image.service.test"),
	}
}

func TestGetSourcePresignedURLWithFiles_UsesFilesNoKalibr(t *testing.T) {
	ctx := context.Background()
	analysisID := "11111111-1111-1111-1111-111111111111"
	stub := &stubAnalysisBucket{}
	svc := testImageServiceWithStub(stub)

	url, err := svc.GetSourcePresignedURLWithFiles(ctx, analysisID, 0, []string{"photo.jpg"})
	if err != nil {
		t.Fatal(err)
	}
	wantKey := imageutil.SourceKey(analysisID, "photo.jpg")
	if stub.lastPresignKey != wantKey {
		t.Fatalf("presign key: got %q want %q", stub.lastPresignKey, wantKey)
	}
	if !strings.Contains(url, wantKey) {
		t.Fatalf("url should contain key: %q", url)
	}
}

func TestGetSourcePresignedURLWithFiles_FallbackProbe(t *testing.T) {
	ctx := context.Background()
	analysisID := "22222222-2222-2222-2222-222222222222"
	fallbackKey := imageutil.SourceKey(analysisID, "image_0.jpg")
	stub := &stubAnalysisBucket{
		streamOKKeys: map[string]struct{}{fallbackKey: {}},
	}
	svc := testImageServiceWithStub(stub)

	_, err := svc.GetSourcePresignedURLWithFiles(ctx, analysisID, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	if stub.lastPresignKey != fallbackKey {
		t.Fatalf("presign key: got %q want %q", stub.lastPresignKey, fallbackKey)
	}
}

func TestGetSourcePresignedURLWithFiles_Validation(t *testing.T) {
	ctx := context.Background()
	svc := testImageServiceWithStub(&stubAnalysisBucket{})
	_, err := svc.GetSourcePresignedURLWithFiles(ctx, "", 0, []string{"a.jpg"})
	if err == nil {
		t.Fatal("expected error for empty analysis id")
	}
	_, err = svc.GetSourcePresignedURLWithFiles(
		ctx,
		"11111111-1111-1111-1111-111111111111",
		-1,
		nil,
	)
	if !errors.Is(err, ErrIndexOutOfRange) {
		t.Fatalf("expected ErrIndexOutOfRange for negative index, got %v", err)
	}
}
