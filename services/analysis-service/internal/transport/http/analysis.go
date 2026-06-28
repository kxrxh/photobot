package http

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"reflect"
	"strconv"
	"strings"
	"time"

	"csort.ru/analysis-service/internal/analysis"
	"csort.ru/analysis-service/internal/api/auth"
	"csort.ru/analysis-service/internal/apierrors"
	"csort.ru/analysis-service/internal/dto"
	"csort.ru/analysis-service/internal/image"
	"csort.ru/analysis-service/internal/middleware"
	"csort.ru/analysis-service/internal/multipartutil"
	"csort.ru/analysis-service/internal/requests"
	"csort.ru/analysis-service/internal/storage"
	"csort.ru/analysis-service/internal/transport/response"
	validatepkg "csort.ru/analysis-service/internal/validator"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

const (
	maxFileSize  = 50 * 1024 * 1024  // 50MB per file limit
	maxTotalSize = 200 * 1024 * 1024 // 200MB total limit
)

type AnalysisHandler struct {
	log               zerolog.Logger
	analysisService   *analysis.Service
	imageService      *image.Service
	requestsService   *requests.Service
	tempStorageClient *storage.Client
	validator         *validator.Validate
}

func NewAnalysisHandler(
	log zerolog.Logger,
	analysisService *analysis.Service,
	imageService *image.Service,
	requestsService *requests.Service,
	tempStorageClient *storage.Client,
	v *validator.Validate,
) *AnalysisHandler {
	return &AnalysisHandler{
		log:               log,
		analysisService:   analysisService,
		imageService:      imageService,
		requestsService:   requestsService,
		tempStorageClient: tempStorageClient,
		validator:         v,
	}
}

func (h *AnalysisHandler) GetAnalyses(c fiber.Ctx) error {
	identity, ok := c.Locals(middleware.UserDataKey).(*auth.Identity)
	if !ok || identity == nil {
		h.log.Warn().Msg("get analyses rejected: auth required")
		return apierrors.Unauthorized("Authentication required")
	}

	var userIDs []int64
	if identity.TelegramID != nil {
		userIDs = append(userIDs, *identity.TelegramID)
	}
	if identity.MaxID != nil {
		userIDs = append(userIDs, *identity.MaxID)
	}
	if len(userIDs) == 0 {
		h.log.Warn().Msg("get analyses rejected: no messenger id")
		return apierrors.BadRequest("user has no messenger id (telegram or max)")
	}

	var queryReq dto.GetAnalysesQueryRequest
	if err := c.Bind().Query(&queryReq); err != nil {
		h.log.Error().Err(err).Msg("get analyses rejected: invalid query params")
		return apierrors.BadRequest("invalid query params")
	}

	queryReq.SetDefaults()
	if err := h.validator.Struct(queryReq); err != nil {
		var validationErrors []validationErrorDetail
		for _, fe := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, validationErrorDetail{
				Field:   validatepkg.GetJSONFieldName(fe.Field(), reflect.TypeOf(queryReq)),
				Message: validatepkg.Translate(fe),
			})
		}
		h.log.Error().Err(err).Msg("get analyses rejected: validation failed")
		return apierrors.WithDetails(apierrors.BadRequest("Validation failed"), validationErrors)
	}

	start := time.Now()
	ctx := c.Context()
	domainResp, err := h.analysisService.List(
		ctx,
		userIDs,
		GetAnalysesQueryRequestToListParams(queryReq),
	)
	if err != nil {
		h.log.Error().Err(err).Interface("user_ids", userIDs).Msg("get analyses failed")
		return err
	}

	resp := dto.PaginatedAnalysesResponse{
		Data:   make([]dto.AnalysisResponse, len(domainResp.Data)),
		Total:  domainResp.Total,
		Limit:  domainResp.Limit,
		Offset: domainResp.Offset,
	}
	for i := range domainResp.Data {
		resp.Data[i] = AnalysisListItemToResponse(domainResp.Data[i])
	}
	enrichAnalysisListFirstSourceURL(ctx, h.imageService, resp.Data)

	h.log.Info().
		Interface("user_ids", userIDs).
		Str("product", queryReq.Product).
		Str("analysis_id_query", queryReq.ID).
		Str("sort_by", queryReq.SortBy).
		Str("sort_order", queryReq.SortOrder).
		Int64("total", domainResp.Total).
		Int("returned", len(resp.Data)).
		Int32("limit", domainResp.Limit).
		Int32("offset", domainResp.Offset).
		Dur("duration", time.Since(start)).
		Msg("analysis flow: list analyses completed")

	return response.OK(c, resp)
}

func (h *AnalysisHandler) GetAnalysisByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return apierrors.New(fiber.StatusBadRequest, "Missing id parameter")
	}

	ctx := c.Context()
	domain, err := h.analysisService.GetByID(ctx, id)
	if err != nil {
		return err
	}

	startEnrich := time.Now()
	resp := AnalysisWithObjectsToResponse(*domain)
	enrichAnalysisWithPresignedURLs(ctx, h.imageService, id, domain, &resp)

	logEvt := h.log.Info().
		Str("analysis_id", id).
		Int64("user_id", domain.Analysis.UserID).
		Int("object_count", len(domain.Objects)).
		Int("files_source", len(domain.FilesSource)).
		Int("files_output", len(domain.FilesOutput)).
		Int("presigned_source_urls", len(resp.FilesSourceURLs)).
		Int("presigned_output_urls", len(resp.FilesOutputURLs)).
		Dur("enrich_presign_duration", time.Since(startEnrich))
	if domain.Product != nil {
		logEvt = logEvt.Str("product", *domain.Product)
	}
	logEvt.Msg("analysis flow: get analysis by id completed")

	return response.OK(c, resp)
}

func enrichAnalysisWithPresignedURLs(
	ctx context.Context,
	imgSvc *image.Service,
	analysisID string,
	domain *analysis.AnalysisWithObjects,
	resp *dto.AnalysisWithObjectsResponse,
) {
	g, ctx := errgroup.WithContext(ctx)

	var sourceURLs []string
	g.Go(func() error {
		sourceURLs = make([]string, 0, len(domain.FilesSource))
		if len(domain.FilesSource) > 0 {
			for i := range domain.FilesSource {
				url, err := imgSvc.GetSourcePresignedURL(ctx, analysisID, i)
				if err != nil {
					continue
				}
				sourceURLs = append(sourceURLs, url)
			}
			return nil
		}
		for i := range 20 {
			url, err := imgSvc.GetSourcePresignedURL(ctx, analysisID, i)
			if err != nil {
				if i == 0 {
					return nil
				}
				break
			}
			sourceURLs = append(sourceURLs, url)
		}
		return nil
	})

	var outputURLs []string
	g.Go(func() error {
		outputURLs = make([]string, 0, len(domain.FilesOutput))
		if len(domain.FilesOutput) > 0 {
			for i := range domain.FilesOutput {
				url, err := imgSvc.GetOutputPresignedURL(ctx, analysisID, i)
				if err != nil {
					continue
				}
				outputURLs = append(outputURLs, url)
			}
			return nil
		}
		for i := range 20 {
			url, err := imgSvc.GetOutputPresignedURL(ctx, analysisID, i)
			if err != nil {
				if i == 0 {
					return nil
				}
				break
			}
			outputURLs = append(outputURLs, url)
		}
		return nil
	})

	g.Go(func() error {
		for i := range resp.Objects {
			obj := domain.Objects[i]
			if obj.File == nil || *obj.File == "" {
				continue
			}
			url, err := imgSvc.GetObjectPresignedURL(ctx, analysisID, *obj.File)
			if err != nil {
				continue
			}
			resp.Objects[i].ImageURL = &url
		}
		return nil
	})

	_ = g.Wait()
	if len(sourceURLs) > 0 {
		resp.FilesSourceURLs = sourceURLs
	}
	if len(outputURLs) > 0 {
		resp.FilesOutputURLs = outputURLs
	}
}

func enrichAnalysisListFirstSourceURL(
	ctx context.Context,
	imgSvc *image.Service,
	items []dto.AnalysisResponse,
) {
	for i := range items {
		id := items[i].ID
		if id == "" {
			continue
		}
		url, err := imgSvc.GetSourcePresignedURLWithFiles(ctx, id, 0, items[i].FilesSource)
		if err != nil {
			continue
		}
		items[i].FilesSourceURLs = []string{url}
	}
}

func (h *AnalysisHandler) GetAnalysisObjects(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return apierrors.New(fiber.StatusBadRequest, "Missing id parameter")
	}

	ctx := c.Context()
	start := time.Now()
	domainObjs, err := h.analysisService.GetObjects(ctx, id)
	if err != nil {
		h.log.Error().Err(err).Str("analysis_id", id).Msg("get analysis objects failed")
		return err
	}

	resp := make([]dto.ObjectResponse, len(domainObjs))
	presigned := 0
	for i := range domainObjs {
		resp[i] = ObjectToResponse(domainObjs[i])
		obj := domainObjs[i]
		if obj.File != nil && *obj.File != "" {
			url, err := h.imageService.GetObjectPresignedURL(ctx, id, *obj.File)
			if err == nil {
				resp[i].ImageURL = &url
				presigned++
			}
		}
	}
	h.log.Info().
		Str("analysis_id", id).
		Int("object_count", len(domainObjs)).
		Int("presigned_image_urls", presigned).
		Dur("duration", time.Since(start)).
		Msg("analysis flow: get analysis objects completed")
	return response.OK(c, resp)
}

func (h *AnalysisHandler) CreateAnalysis(c fiber.Ctx) error {
	identity, ok := c.Locals(middleware.UserDataKey).(*auth.Identity)
	if !ok || identity == nil {
		h.log.Warn().Msg("create analysis rejected: auth required")
		return apierrors.BadRequest("user identity is required")
	}

	var userID string
	var platform string
	switch {
	case identity.MaxID != nil:
		platform = "max"
		userID = strconv.FormatInt(*identity.MaxID, 10)
	case identity.TelegramID != nil:
		platform = "telegram"
		userID = strconv.FormatInt(*identity.TelegramID, 10)
	default:
		return apierrors.BadRequest("user has no messenger id (telegram or max)")
	}

	product := strings.TrimSpace(c.FormValue("product"))
	bot := strings.TrimSpace(c.FormValue("bot"))
	massLiterStr := strings.TrimSpace(c.FormValue("mass_liter"))
	location := strings.TrimSpace(c.FormValue("location"))
	yearStr := strings.TrimSpace(c.FormValue("year"))
	mass1000Str := strings.TrimSpace(c.FormValue("mass_1000"))
	massStr := strings.TrimSpace(c.FormValue("mass"))

	var mass1000 *float64
	if mass1000Str != "" {
		val, err := strconv.ParseFloat(mass1000Str, 64)
		if err != nil {
			h.log.Error().
				Err(err).
				Str("mass_1000", mass1000Str).
				Msg("create analysis rejected: invalid mass_1000")
			return apierrors.BadRequest("invalid mass_1000 value")
		}
		mass1000 = &val
	}

	var mass *float64
	if massStr != "" {
		val, err := strconv.ParseFloat(massStr, 64)
		if err != nil {
			h.log.Error().
				Err(err).
				Str("mass", massStr).
				Msg("create analysis rejected: invalid mass")
			return apierrors.BadRequest("invalid mass value")
		}
		mass = &val
	}

	if mass1000 != nil && mass != nil {
		return apierrors.BadRequest("only one of mass_1000 or mass can be provided")
	}

	var massLiter *float64
	if massLiterStr != "" {
		val, err := strconv.ParseFloat(massLiterStr, 64)
		if err != nil {
			h.log.Error().
				Err(err).
				Str("mass_liter", massLiterStr).
				Msg("create analysis rejected: invalid mass_liter")
			return apierrors.BadRequest("invalid mass_liter value")
		}
		massLiter = &val
	}

	if yearStr == "" {
		yearStr = strconv.Itoa(time.Now().Year())
	}

	fields := dto.CreateAnalysisFormFields{
		Product:  product,
		Bot:      bot,
		Location: location,
		Year:     yearStr,
	}
	if err := h.validator.Struct(fields); err != nil {
		var validationErrors []validationErrorDetail
		for _, fe := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, validationErrorDetail{
				Field:   validatepkg.GetJSONFieldName(fe.Field(), reflect.TypeOf(fields)),
				Message: validatepkg.Translate(fe),
			})
		}
		h.log.Error().Err(err).Msg("create analysis rejected: validation failed")
		return apierrors.WithDetails(apierrors.BadRequest("Validation failed"), validationErrors)
	}

	flowStart := time.Now()

	files, err := h.processUploadedFiles(c)
	if err != nil {
		h.log.Error().Err(err).Msg("create analysis rejected: process files failed")
		return apierrors.BadRequest(err.Error())
	}

	requestID, err := h.requestsService.Create(c.Context(), &requests.CreateParams{
		Product:   product,
		UserID:    userID,
		Platform:  platform,
		Bot:       bot,
		Year:      yearStr,
		MassLiter: massLiter,
		Location:  location,
		Mass1000:  mass1000,
		Mass:      mass,
		Images:    files,
	})
	if err != nil {
		h.tempStorageClient.DeleteFiles(c.Context(), imageIDsFrom(files))
		h.log.Error().Err(err).Msg("create analysis failed")
		return apierrors.BadGatewayWrap(err, err.Error())
	}

	h.log.Info().
		Str("request_id", requestID).
		Str("product", product).
		Str("platform", platform).
		Str("user_id", userID).
		Str("bot", bot).
		Str("year", yearStr).
		Int("image_count", len(files)).
		Dur("duration", time.Since(flowStart)).
		Msg("analysis flow: create analysis request persisted")

	return response.OK(c, dto.CreateAnalysisResponse{RequestID: requestID})
}

func (h *AnalysisHandler) MergeAnalyses(c fiber.Ctx) error {
	var req dto.MergeAnalysesRequest
	if err := c.Bind().Body(&req); err != nil {
		h.log.Error().Err(err).Msg("merge analyses rejected: invalid payload")
		return apierrors.BadRequest("invalid payload")
	}

	identity, ok := c.Locals(middleware.UserDataKey).(*auth.Identity)
	if ok && identity != nil {
		if identity.MaxID != nil {
			req.UserID = *identity.MaxID
		} else if identity.TelegramID != nil {
			req.UserID = *identity.TelegramID
		}
	}

	if req.UserID <= 0 {
		h.log.Error().Msg("merge analyses rejected: invalid user_id")
		return apierrors.BadRequest("user_id is required and must be positive")
	}

	if err := h.validator.Struct(req); err != nil {
		var validationErrors []validationErrorDetail
		for _, fe := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, validationErrorDetail{
				Field:   validatepkg.GetJSONFieldName(fe.Field(), reflect.TypeOf(req)),
				Message: validatepkg.Translate(fe),
			})
		}
		h.log.Error().Err(err).Msg("merge analyses rejected: validation failed")
		return apierrors.WithDetails(apierrors.BadRequest("Validation failed"), validationErrors)
	}

	if err := h.analysisService.Merge(c.Context(), req.UserID, req.Analyses); err != nil {
		h.log.Error().Err(err).Msg("merge analyses failed")
		return apierrors.BadGatewayWrap(err, err.Error())
	}

	nMerge := len(req.Analyses)
	idLimit := nMerge
	if idLimit > 12 {
		idLimit = 12
	}
	h.log.Info().
		Int64("user_id", req.UserID).
		Int("merge_count", nMerge).
		Strs("analysis_ids_sample", req.Analyses[:idLimit]).
		Msg("analysis flow: merge analyses completed")

	return response.OK(c, dto.MergeAnalysesResponse{Message: "Analyses merged successfully"})
}

func (h *AnalysisHandler) processUploadedFiles(
	c fiber.Ctx,
) ([]requests.CreateAnalysisImageFile, error) {
	multipartForm, err := c.MultipartForm()
	if err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	fileHeaders := multipartutil.CollectFileHeaders(
		multipartForm,
		multipartutil.DefaultFormFileKeys,
	)
	if len(fileHeaders) == 0 {
		h.log.Warn().
			Strs("form_file_keys", multipartutil.FormFileKeys(multipartForm)).
			Msg("create analysis rejected: no files")
		return nil, errors.New("at least one file is required (use form field 'files' for uploads)")
	}

	var totalSize int64
	for _, fileHeader := range fileHeaders {
		if err := h.validateFileHeader(fileHeader, &totalSize); err != nil {
			return nil, err
		}
	}

	results := make([]requests.CreateAnalysisImageFile, len(fileHeaders))
	g, ctx := errgroup.WithContext(c.Context())
	for i, fileHeader := range fileHeaders {
		i, fileHeader := i, fileHeader
		g.Go(func() error {
			uploadResult, err := h.tempStorageClient.UploadMultipartFile(ctx, fileHeader)
			if err != nil {
				return fmt.Errorf("failed to upload file %s: %w", fileHeader.Filename, err)
			}
			results[i] = requests.CreateAnalysisImageFile{
				ID:       uploadResult.FileID,
				ImageURL: uploadResult.URL,
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		var uploadedIDs []string
		for _, r := range results {
			if r.ID != "" {
				uploadedIDs = append(uploadedIDs, r.ID)
			}
		}
		if len(uploadedIDs) > 0 {
			h.tempStorageClient.DeleteFiles(c.Context(), uploadedIDs)
		}
		return nil, err
	}

	return results, nil
}

func imageIDsFrom(files []requests.CreateAnalysisImageFile) []string {
	ids := make([]string, 0, len(files))
	for _, f := range files {
		ids = append(ids, f.ID)
	}
	return ids
}

func (h *AnalysisHandler) validateFileHeader(
	fileHeader *multipart.FileHeader,
	totalSize *int64,
) error {
	if fileHeader.Size == 0 {
		return fmt.Errorf("file %s is empty", fileHeader.Filename)
	}
	if fileHeader.Size > maxFileSize {
		return fmt.Errorf(
			"file %s is too large (max %d MB)",
			fileHeader.Filename,
			maxFileSize/(1024*1024),
		)
	}
	*totalSize += fileHeader.Size
	if *totalSize > maxTotalSize {
		return fmt.Errorf("total files size exceeds limit of %d MB", maxTotalSize/(1024*1024))
	}
	return nil
}
