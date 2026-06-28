package http

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"csort.ru/analysis-service/internal/analysis"
	"csort.ru/analysis-service/internal/config"
	"csort.ru/analysis-service/internal/objects"
	"csort.ru/analysis-service/internal/sharelink"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type stubImageService struct {
	objects map[int32]struct {
		body     []byte
		mimeType string
		err      error
	}
}

func (s *stubImageService) GetSourceStream(
	_ context.Context,
	_ string,
	_ int,
) (io.ReadCloser, string, error) {
	return nil, "", errors.New("not implemented")
}

func (s *stubImageService) GetOutputStream(
	_ context.Context,
	_ string,
	_ int,
) (io.ReadCloser, string, error) {
	return nil, "", errors.New("not implemented")
}

func (s *stubImageService) GetObjectStream(
	_ context.Context,
	_ string,
	objectID int32,
) (io.ReadCloser, string, error) {
	if obj, ok := s.objects[objectID]; ok {
		if obj.err != nil {
			return nil, "", obj.err
		}
		return io.NopCloser(bytes.NewReader(obj.body)), obj.mimeType, nil
	}
	return nil, "", errors.New("not found")
}

func (s *stubImageService) GetObjectStreamByFile(
	_ context.Context,
	_ string,
	objectFile string,
) (io.ReadCloser, string, error) {
	for objectID, obj := range s.objects {
		if obj.err != nil {
			continue
		}
		expectedFile := ""
		switch objectID {
		case 0:
			expectedFile = "a.png"
		case 1:
			expectedFile = "b.jpg"
		default:
			expectedFile = ""
		}
		if expectedFile == objectFile {
			return io.NopCloser(bytes.NewReader(obj.body)), obj.mimeType, nil
		}
	}
	return nil, "", errors.New("not found")
}

type stubAnalysisLookup struct {
	domain *analysis.AnalysisWithObjects
	err    error
}

func (s *stubAnalysisLookup) GetByID(
	_ context.Context,
	_ string,
) (*analysis.AnalysisWithObjects, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.domain, nil
}

func ptrString(v string) *string { return &v }

func makePackAuth(secret string) *ReportPackAuthorizer {
	return NewReportPackAuthorizer(
		config.ShareLinkConfig{
			HMACSecret:      secret,
			MaxClockSkewSec: 60,
		},
		nil,
		nil,
	)
}

func signedQuery(secret, analysisID string) string {
	exp := time.Now().Unix() + 3600
	sig := sharelink.Sign([]byte(secret), analysisID, exp)
	return fmt.Sprintf("?exp=%d&sig=%s", exp, sig)
}

func TestDownloadObjectImagesArchive_UnauthorizedWithoutAuth(t *testing.T) {
	app := fiber.New()
	handler := NewImageHandler(
		zerolog.Nop(),
		&stubImageService{},
		&stubAnalysisLookup{},
		makePackAuth("archive-secret"),
	)
	app.Get("/analyses/:id/images/objects/archive", handler.DownloadObjectImagesArchive)

	req := httptest.NewRequest(nethttp.MethodGet, "/analyses/a1/images/objects/archive", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusUnauthorized {
		t.Fatalf("status code: got %d want %d", resp.StatusCode, fiber.StatusUnauthorized)
	}
}

func TestDownloadObjectImagesArchive_WithShareLinkReturnsZip(t *testing.T) {
	const analysisID = "a-zip"
	const secret = "archive-secret"
	app := fiber.New()
	handler := NewImageHandler(
		zerolog.Nop(),
		&stubImageService{
			objects: map[int32]struct {
				body     []byte
				mimeType string
				err      error
			}{
				0: {body: []byte("image-0"), mimeType: "image/png"},
				1: {body: []byte("image-1"), mimeType: "image/jpeg"},
			},
		},
		&stubAnalysisLookup{
			domain: &analysis.AnalysisWithObjects{
				Objects: []objects.Object{
					{ObjectMetadata: objects.ObjectMetadata{ID: 0}, File: ptrString("a.png")},
					{ObjectMetadata: objects.ObjectMetadata{ID: 1}, File: ptrString("b.jpg")},
				},
			},
		},
		makePackAuth(secret),
	)
	app.Get("/analyses/:id/images/objects/archive", handler.DownloadObjectImagesArchive)

	req := httptest.NewRequest(
		nethttp.MethodGet,
		"/analyses/"+analysisID+"/images/objects/archive"+signedQuery(secret, analysisID),
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("status code: got %d want %d", resp.StatusCode, fiber.StatusOK)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	zipReader, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatalf("invalid zip: %v", err)
	}
	if len(zipReader.File) != 2 {
		t.Fatalf("zip entries: got %d want 2", len(zipReader.File))
	}
	if zipReader.File[0].Name != "object_0.png" {
		t.Fatalf("first entry name: got %q want %q", zipReader.File[0].Name, "object_0.png")
	}
	if zipReader.File[1].Name != "object_1.jpg" {
		t.Fatalf("second entry name: got %q want %q", zipReader.File[1].Name, "object_1.jpg")
	}
}

func TestDownloadObjectImagesArchive_NoObjectFiles(t *testing.T) {
	const analysisID = "a-empty"
	const secret = "archive-secret"
	app := fiber.New()
	handler := NewImageHandler(
		zerolog.Nop(),
		&stubImageService{},
		&stubAnalysisLookup{
			domain: &analysis.AnalysisWithObjects{
				Objects: []objects.Object{
					{ObjectMetadata: objects.ObjectMetadata{ID: 0}, File: nil},
				},
			},
		},
		makePackAuth(secret),
	)
	app.Get("/analyses/:id/images/objects/archive", handler.DownloadObjectImagesArchive)

	req := httptest.NewRequest(
		nethttp.MethodGet,
		"/analyses/"+analysisID+"/images/objects/archive"+signedQuery(secret, analysisID),
		nil,
	)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusNotFound {
		t.Fatalf("status code: got %d want %d", resp.StatusCode, fiber.StatusNotFound)
	}
}
