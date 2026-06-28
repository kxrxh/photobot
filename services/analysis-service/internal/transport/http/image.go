package http

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"csort.ru/analysis-service/internal/analysis"
	"csort.ru/analysis-service/internal/apierrors"
	apifiber "csort.ru/analysis-service/internal/apierrors/fiber"
	"csort.ru/analysis-service/internal/image"
	"csort.ru/analysis-service/internal/imageutil"
	"csort.ru/analysis-service/internal/objects"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type ImageHandler struct {
	log             zerolog.Logger
	service         imageStreamService
	analysisService analysisLookupService
	packAuth        *ReportPackAuthorizer
}

type imageStreamService interface {
	GetSourceStream(
		ctx context.Context,
		analysisID string,
		index int,
	) (io.ReadCloser, string, error)
	GetOutputStream(
		ctx context.Context,
		analysisID string,
		index int,
	) (io.ReadCloser, string, error)
	GetObjectStream(
		ctx context.Context,
		analysisID string,
		objectID int32,
	) (io.ReadCloser, string, error)
	GetObjectStreamByFile(
		ctx context.Context,
		analysisID string,
		objectFile string,
	) (io.ReadCloser, string, error)
}

type analysisLookupService interface {
	GetByID(ctx context.Context, analysisID string) (*analysis.AnalysisWithObjects, error)
}

type archiveFetchJob struct {
	index  int
	object objects.Object
}

type archiveFetchResult struct {
	index    int
	object   objects.Object
	mimeType string
	body     io.ReadCloser
	err      error
}

func NewImageHandler(
	log zerolog.Logger,
	service imageStreamService,
	analysisService analysisLookupService,
	packAuth *ReportPackAuthorizer,
) *ImageHandler {
	return &ImageHandler{
		log:             log,
		service:         service,
		analysisService: analysisService,
		packAuth:        packAuth,
	}
}

func (h *ImageHandler) DownloadSourceImage(c fiber.Ctx) error {
	analysisID := c.Params("id")
	indexStr := c.Params("index")

	if analysisID == "" {
		h.log.Warn().Msg("download source image rejected: missing analysis_id")
		return apierrors.BadRequest("analysis ID is required")
	}

	if err := h.packAuth.EnsureReportPackAccess(c, analysisID); err != nil {
		return err
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 {
		h.log.Warn().
			Str("analysis_id", analysisID).
			Str("index", indexStr).
			Msg("download source image rejected: invalid index")
		return apierrors.BadRequest("invalid image index")
	}

	rc, mimeType, err := h.service.GetSourceStream(c.Context(), analysisID, index)
	if err != nil {
		return h.handleImageError(analysisID, index, err)
	}
	defer func() { _ = rc.Close() }()

	return h.sendImageStream(c, rc, mimeType, fmt.Sprintf("source_image_%d", index+1))
}

func (h *ImageHandler) DownloadOutputImage(c fiber.Ctx) error {
	analysisID := c.Params("id")
	indexStr := c.Params("index")

	if analysisID == "" {
		h.log.Warn().Msg("download output image rejected: missing analysis_id")
		return apierrors.BadRequest("analysis ID is required")
	}

	if err := h.packAuth.EnsureReportPackAccess(c, analysisID); err != nil {
		return err
	}

	index, err := strconv.Atoi(indexStr)
	if err != nil || index < 0 {
		h.log.Warn().
			Str("analysis_id", analysisID).
			Str("index", indexStr).
			Msg("download output image rejected: invalid index")
		return apierrors.BadRequest("invalid image index")
	}

	rc, mimeType, err := h.service.GetOutputStream(c.Context(), analysisID, index)
	if err != nil {
		return h.handleImageError(analysisID, index, err)
	}
	defer func() { _ = rc.Close() }()

	return h.sendImageStream(c, rc, mimeType, fmt.Sprintf("analysis_image_%d", index+1))
}

func (h *ImageHandler) DownloadObjectImage(c fiber.Ctx) error {
	analysisID := c.Params("id")
	objectIDStr := c.Params("objectId")

	if analysisID == "" {
		h.log.Warn().Msg("download object image rejected: missing analysis_id")
		return apierrors.BadRequest("analysis ID is required")
	}

	if err := h.packAuth.EnsureReportPackAccess(c, analysisID); err != nil {
		return err
	}

	objectID, err := strconv.ParseInt(objectIDStr, 10, 32)
	if err != nil {
		h.log.Warn().
			Str("analysis_id", analysisID).
			Str("object_id", objectIDStr).
			Msg("download object image rejected: invalid object_id")
		return apierrors.BadRequest("invalid object ID")
	}

	rc, mimeType, err := h.service.GetObjectStream(c.Context(), analysisID, int32(objectID))
	if err != nil {
		return h.handleImageError(analysisID, int(objectID), err)
	}
	defer func() { _ = rc.Close() }()

	return h.sendImageStream(c, rc, mimeType, fmt.Sprintf("object_%d", objectID))
}

const archiveMaxWorkers = 8

func archiveFetchWorkerCount(objectCount int) int {
	if objectCount <= 0 {
		return 1
	}
	return min(archiveMaxWorkers, objectCount)
}

func (h *ImageHandler) DownloadObjectImagesArchive(c fiber.Ctx) error {
	analysisID := c.Params("id")
	if analysisID == "" {
		h.log.Warn().Msg("download objects archive rejected: missing analysis_id")
		return apierrors.BadRequest("analysis ID is required")
	}

	if err := h.packAuth.EnsureReportPackAccess(c, analysisID); err != nil {
		return apifiber.ErrorHandler(c, err)
	}

	domain, err := h.analysisService.GetByID(c.Context(), analysisID)
	if err != nil {
		return apifiber.ErrorHandler(c, err)
	}

	candidates := make([]objects.Object, 0, len(domain.Objects))
	for _, obj := range domain.Objects {
		if obj.File == nil || strings.TrimSpace(*obj.File) == "" {
			continue
		}
		candidates = append(candidates, obj)
	}
	if len(candidates) == 0 {
		return apifiber.ErrorHandler(c, apierrors.NotFound("object images not found"))
	}

	c.Set("Content-Type", "application/zip")
	c.Set(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s_objects.zip"`, sanitizeFilename(analysisID)),
	)
	c.Set("Cache-Control", "no-cache")

	zipWriter := zip.NewWriter(c.Response().BodyWriter())
	added := 0

	workers := archiveFetchWorkerCount(len(candidates))
	jobs := make(chan archiveFetchJob)
	resultsBuf := min(len(candidates), min(128, max(48, workers*2)))
	results := make(chan archiveFetchResult, resultsBuf)
	var wg sync.WaitGroup

	for range workers {
		wg.Go(func() {
			for job := range jobs {
				results <- h.fetchArchiveObject(c.Context(), analysisID, job)
			}
		})
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	go func() {
		for i, obj := range candidates {
			jobs <- archiveFetchJob{index: i, object: obj}
		}
		close(jobs)
	}()

	pending := make(map[int]archiveFetchResult, min(len(candidates), 128))
	nextIndex := 0
	for result := range results {
		pending[result.index] = result
		for {
			ready, ok := pending[nextIndex]
			if !ok {
				break
			}
			delete(pending, nextIndex)
			nextIndex++

			if ready.err != nil {
				h.log.Warn().
					Err(ready.err).
					Str("analysis_id", analysisID).
					Int32("object_id", ready.object.ID).
					Msg("failed to fetch object image for archive")
				continue
			}

			ok, addErr := h.addObjectStreamToZip(
				zipWriter,
				ready.object,
				ready.body,
				ready.mimeType,
			)
			if addErr != nil {
				h.log.Warn().
					Err(addErr).
					Str("analysis_id", analysisID).
					Int32("object_id", ready.object.ID).
					Msg("failed to add object image to archive")
				continue
			}
			if ok {
				added++
			}
		}
	}

	if err := zipWriter.Close(); err != nil {
		return apifiber.ErrorHandler(
			c,
			apierrors.InternalWrap(err, "failed to finalize object images archive"),
		)
	}
	if added == 0 {
		return apifiber.ErrorHandler(c, apierrors.NotFound("object images not found"))
	}
	return nil
}

func (h *ImageHandler) fetchArchiveObject(
	ctx context.Context,
	analysisID string,
	job archiveFetchJob,
) archiveFetchResult {
	rc, mimeType, err := h.service.GetObjectStreamByFile(ctx, analysisID, *job.object.File)
	if err != nil {
		return archiveFetchResult{
			index:  job.index,
			object: job.object,
			err:    err,
		}
	}

	return archiveFetchResult{
		index:    job.index,
		object:   job.object,
		mimeType: mimeType,
		body:     rc,
	}
}

func (h *ImageHandler) addObjectStreamToZip(
	zipWriter *zip.Writer,
	object objects.Object,
	rc io.ReadCloser,
	mimeType string,
) (bool, error) {
	if rc == nil {
		return false, nil
	}
	defer func() { _ = rc.Close() }()

	if object.File == nil || strings.TrimSpace(*object.File) == "" {
		return false, nil
	}
	ext := objectFileExtension(*object.File, mimeType)
	entryName := fmt.Sprintf("object_%d.%s", object.ID, ext)
	header := &zip.FileHeader{
		Name:   entryName,
		Method: zip.Store,
	}
	entryWriter, err := zipWriter.CreateHeader(header)
	if err != nil {
		return false, err
	}
	if _, err := io.Copy(entryWriter, rc); err != nil {
		return false, err
	}
	return true, nil
}

func objectFileExtension(objectFile string, mimeType string) string {
	ext := strings.TrimPrefix(strings.ToLower(filepath.Ext(strings.TrimSpace(objectFile))), ".")
	if ext != "" {
		return ext
	}
	return getExtensionFromMimeType(mimeType)
}

func (h *ImageHandler) handleImageError(
	analysisID string,
	index int,
	err error,
) error {
	if errors.Is(err, image.ErrNotFound) || errors.Is(err, image.ErrIndexOutOfRange) ||
		errors.Is(err, imageutil.ErrIndexOutOfRange) {
		return apierrors.NotFound("not found")
	}
	h.log.Error().
		Err(err).
		Str("analysis_id", analysisID).
		Int("index", index).
		Msg("get image failed")
	return apierrors.Internal("failed to retrieve image")
}

func (h *ImageHandler) sendImageStream(
	c fiber.Ctx,
	rc io.Reader,
	mimeType string,
	baseFilename string,
) error {
	ext := getExtensionFromMimeType(mimeType)
	filename := fmt.Sprintf("%s.%s", baseFilename, ext)

	c.Set("Content-Type", mimeType)
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Set("Cache-Control", "no-cache")

	_, err := io.Copy(c.Response().BodyWriter(), rc)
	return err
}

func getExtensionFromMimeType(mimeType string) string {
	if strings.Contains(mimeType, "png") {
		return "png"
	}
	if strings.Contains(mimeType, "jpeg") || strings.Contains(mimeType, "jpg") {
		return "jpg"
	}
	if strings.Contains(mimeType, "gif") {
		return "gif"
	}
	if strings.Contains(mimeType, "webp") {
		return "webp"
	}
	if strings.Contains(mimeType, "heic") {
		return "heic"
	}
	return "png"
}
