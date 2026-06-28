package http

import (
	"fmt"
	"strconv"

	"csort.ru/coffeebot/internal/dto"
	"csort.ru/coffeebot/internal/transport"
	"csort.ru/coffeebot/internal/transport/response"
	"csort.ru/coffeebot/internal/weed"
	"csort.ru/coffeebot/internal/weed/analysis"
	"csort.ru/coffeebot/internal/weed/image"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type WeedHandler struct {
	weedService         *weed.Service
	weedImageService    *image.Service
	weedAnalysesService *analysis.Service
	validator           *validator.Validate
	log                 zerolog.Logger
}

func NewWeedHandler(
	weedService *weed.Service,
	weedImageService *image.Service,
	weedAnalysesService *analysis.Service,
	zlog zerolog.Logger,
) *WeedHandler {
	return &WeedHandler{
		weedService:         weedService,
		weedImageService:    weedImageService,
		weedAnalysesService: weedAnalysesService,
		validator:           validator.New(),
		log:                 zlog,
	}
}

func (h *WeedHandler) ListWeeds(c fiber.Ctx) error {
	limit, offset := transport.ClampPaginationFromQuery(
		c.Query("limit", "10"),
		c.Query("offset", "0"),
	)

	sortOrder := c.Query("sort_order", "desc")
	isQuarantineParam := c.Query("is_quarantine")
	var isQuarantine *bool
	if isQuarantineParam != "" {
		v := isQuarantineParam == "true"
		isQuarantine = &v
	}

	params := weed.ListWeedsParams{
		PaginatedRequest: dto.PaginatedRequest{
			Limit:  limit,
			Offset: offset,
		},
		Name:         c.Query("name"),
		MainGroup:    c.Query("main_group"),
		MainSubgroup: c.Query("main_subgroup"),
		Subgroup:     c.Query("subgroup"),
		IsQuarantine: isQuarantine,
		SortOrder:    sortOrder,
	}

	if v := c.Query("l_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.LMin = &f32
		}
	}
	if v := c.Query("l_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.LMax = &f32
		}
	}
	if v := c.Query("w_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.WMin = &f32
		}
	}
	if v := c.Query("w_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.WMax = &f32
		}
	}
	if v := c.Query("lw_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.LWMin = &f32
		}
	}
	if v := c.Query("lw_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.LWMax = &f32
		}
	}
	if v := c.Query("h_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.HMin = &f32
		}
	}
	if v := c.Query("h_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.HMax = &f32
		}
	}
	if v := c.Query("s_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.SMin = &f32
		}
	}
	if v := c.Query("s_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.SMax = &f32
		}
	}
	if v := c.Query("v_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.VMin = &f32
		}
	}
	if v := c.Query("v_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.VMax = &f32
		}
	}
	if v := c.Query("r_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.RMin = &f32
		}
	}
	if v := c.Query("r_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.RMax = &f32
		}
	}
	if v := c.Query("g_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.GMin = &f32
		}
	}
	if v := c.Query("g_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.GMax = &f32
		}
	}
	if v := c.Query("b_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.BMin = &f32
		}
	}
	if v := c.Query("b_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.BMax = &f32
		}
	}
	if v := c.Query("brt_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.BrtMin = &f32
		}
	}
	if v := c.Query("brt_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.BrtMax = &f32
		}
	}
	if v := c.Query("sq_sqcrl_min"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.SqSqcrlMin = &f32
		}
	}
	if v := c.Query("sq_sqcrl_max"); v != "" {
		if f, err := strconv.ParseFloat(v, 32); err == nil {
			f32 := float32(f)
			params.SqSqcrlMax = &f32
		}
	}

	listResult, err := h.weedService.ListWeeds(c.Context(), params)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to list weeds")
		return response.Fail(
			c,
			fiber.StatusInternalServerError,
			"Failed to retrieve weed list",
			nil,
		)
	}
	return response.OK(c, listResult)
}

func (h *WeedHandler) GetWeed(c fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		h.log.Error().Msg("Missing id parameter")
		return response.Fail(c, fiber.StatusBadRequest, "Missing id parameter", nil)
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to parse ID parameter")
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			fmt.Sprintf("Invalid id parameter: %v", err),
			nil,
		)
	}
	weed, err := h.weedService.GetWeedByID(c.Context(), int32(id))
	if err != nil {
		h.log.Error().Err(err).Int32("weed_id", int32(id)).Msg("Failed to get weed")
		return response.Fail(c, fiber.StatusNotFound, "Weed not found", nil)
	}
	return response.OK(c, weed)
}

func (h *WeedHandler) GetFullWeedDetails(c fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		h.log.Error().Msg("Missing id parameter")
		return response.Fail(c, fiber.StatusBadRequest, "Missing id parameter", nil)
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		h.log.Error().Err(err).Msg("Failed to parse ID parameter")
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			fmt.Sprintf("Invalid id parameter: %v", err),
			nil,
		)
	}

	d, err := h.weedService.GetWeedWithDetails(c.Context(), int32(id))
	if err != nil {
		h.log.Error().
			Err(err).
			Int32("weed_id", int32(id)).
			Msg("Failed to get weed with details")
		return response.Fail(c, fiber.StatusNotFound, "Weed not found", nil)
	}
	return response.OK(c, d)
}

func (h *WeedHandler) UpdateWeed(c fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return response.Fail(c, fiber.StatusBadRequest, "Missing id parameter", nil)
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			fmt.Sprintf("Invalid id parameter: %v", err),
			nil,
		)
	}
	var params weed.SaveWeedParams
	if err := c.Bind().Body(&params); err != nil {
		h.log.Error().
			Err(err).
			Int32("weed_id", int32(id)).
			Msg("Failed to parse update weed request")
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", nil)
	}
	if err := h.validator.Struct(params); err != nil {
		h.log.Error().
			Err(err).
			Int32("weed_id", int32(id)).
			Msg("Weed update validation failed")
		return response.Fail(c, fiber.StatusBadRequest, err.Error(), nil)
	}
	weed, err := h.weedService.UpdateWeed(c.Context(), int32(id), params)
	if err != nil {
		h.log.Error().Err(err).Int32("weed_id", int32(id)).Msg("Failed to update weed")
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to update weed", nil)
	}
	return response.OK(c, weed)
}

func (h *WeedHandler) DeleteWeed(c fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return response.Fail(c, fiber.StatusBadRequest, "Missing id parameter", nil)
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			fmt.Sprintf("Invalid id parameter: %v", err),
			nil,
		)
	}
	if err := h.weedService.DeleteWeed(c.Context(), int32(id)); err != nil {
		h.log.Error().Err(err).Int32("weed_id", int32(id)).Msg("Failed to delete weed")
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to delete weed", nil)
	}
	return response.NoContent(c)
}

func (h *WeedHandler) GetWeedImages(c fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return response.Fail(c, fiber.StatusBadRequest, "Missing id parameter", nil)
	}

	id, err := strconv.ParseInt(idStr, 10, 32)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			fmt.Sprintf("Invalid id parameter: %v", err),
			nil,
		)
	}
	images, err := h.weedImageService.GetWeedImages(c.Context(), int32(id))
	if err != nil {
		h.log.Error().Err(err).Int32("weed_id", int32(id)).Msg("Failed to get weed images")
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to retrieve images", nil)
	}
	return response.OK(c, images)
}

func (h *WeedHandler) AddWeedImage(c fiber.Ctx) error {
	weedIDStr := c.Params("id")
	if weedIDStr == "" {
		return response.Fail(c, fiber.StatusBadRequest, "Missing id parameter", nil)
	}

	weedID, err := strconv.ParseInt(weedIDStr, 10, 32)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			fmt.Sprintf("Invalid id parameter: %v", err),
			nil,
		)
	}
	file, err := c.FormFile("file")
	if err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Failed to get file", nil)
	}
	fc, err := file.Open()
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to open file", nil)
	}
	defer func() { _ = fc.Close() }()
	img, err := h.weedImageService.UploadAndAddWeedImage(
		c.Context(),
		int32(weedID),
		fc,
		file.Size,
		file.Header.Get("Content-Type"),
		file.Filename,
	)
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to add image", nil)
	}
	return response.Created(c, img)
}

func (h *WeedHandler) SetPrimaryImage(c fiber.Ctx) error {
	weedIDStr := c.Params("id")
	if weedIDStr == "" {
		return response.Fail(c, fiber.StatusBadRequest, "Missing id parameter", nil)
	}

	weedID, err := strconv.ParseInt(weedIDStr, 10, 32)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			fmt.Sprintf("Invalid id parameter: %v", err),
			nil,
		)
	}

	imageIDStr := c.Params("imageId")
	if imageIDStr == "" {
		return response.Fail(c, fiber.StatusBadRequest, "Missing imageId parameter", nil)
	}

	imageID, err := strconv.ParseInt(imageIDStr, 10, 32)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			fmt.Sprintf("Invalid imageId parameter: %v", err),
			nil,
		)
	}
	if err := h.weedImageService.SetPrimaryImage(
		c.Context(),
		int32(weedID),
		int32(imageID),
	); err != nil {
		h.log.Error().
			Err(err).
			Int32("weed_id", int32(weedID)).
			Int32("image_id", int32(imageID)).
			Msg("Failed to set primary image")
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to set primary image", nil)
	}
	return response.NoContent(c)
}

func (h *WeedHandler) DeleteWeedImage(c fiber.Ctx) error {
	weedIDStr := c.Params("id")
	if weedIDStr == "" {
		return response.Fail(c, fiber.StatusBadRequest, "Missing id parameter", nil)
	}

	weedID, err := strconv.ParseInt(weedIDStr, 10, 32)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			fmt.Sprintf("Invalid id parameter: %v", err),
			nil,
		)
	}

	imageIDStr := c.Params("imageId")
	if imageIDStr == "" {
		return response.Fail(c, fiber.StatusBadRequest, "Missing imageId parameter", nil)
	}

	imageID, err := strconv.ParseInt(imageIDStr, 10, 32)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			fmt.Sprintf("Invalid imageId parameter: %v", err),
			nil,
		)
	}
	if err := h.weedImageService.DeleteWeedImage(c.Context(), int32(imageID)); err != nil {
		h.log.Error().
			Err(err).
			Int32("weed_id", int32(weedID)).
			Int32("image_id", int32(imageID)).
			Msg("Failed to delete weed image")
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to delete image", nil)
	}
	return response.NoContent(c)
}

func (h *WeedHandler) GetWeedAnalyses(c fiber.Ctx) error {
	weedIDStr := c.Params("id")
	if weedIDStr == "" {
		return response.Fail(c, fiber.StatusBadRequest, "Missing id parameter", nil)
	}

	weedID, err := strconv.ParseInt(weedIDStr, 10, 32)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			fmt.Sprintf("Invalid id parameter: %v", err),
			nil,
		)
	}
	analyses, err := h.weedAnalysesService.GetWeedAnalyses(c.Context(), int32(weedID))
	if err != nil {
		return response.Fail(c, fiber.StatusInternalServerError, "Failed to retrieve analyses", nil)
	}
	return response.OK(c, analyses)
}

func (h *WeedHandler) GetWeedAnalysisObjects(c fiber.Ctx) error {
	weedIDStr := c.Params("id")
	if weedIDStr == "" {
		return response.Fail(c, fiber.StatusBadRequest, "Missing id parameter", nil)
	}

	weedID, err := strconv.ParseInt(weedIDStr, 10, 32)
	if err != nil {
		return response.Fail(
			c,
			fiber.StatusBadRequest,
			fmt.Sprintf("Invalid id parameter: %v", err),
			nil,
		)
	}

	weedRow, err := h.weedService.GetWeedWithDetails(c.Context(), int32(weedID))
	if err != nil {
		h.log.Error().
			Err(err).
			Int32("weed_id", int32(weedID)).
			Msg("Failed to get weed analysis objects")
		return response.Fail(
			c,
			fiber.StatusInternalServerError,
			"Failed to retrieve analysis objects",
			nil,
		)
	}

	var excluded []int64
	if weedRow.Statistics != nil {
		excluded = weedRow.Statistics.ExcludedObjects
	} else {
		excluded = []int64{}
	}

	return response.OK(c, dto.WeedAnalysisObject{
		ID:              weedRow.ID,
		WeedID:          weedRow.ID,
		AnalysesIds:     weedRow.Analyses,
		ExcludedObjects: excluded,
	})
}
