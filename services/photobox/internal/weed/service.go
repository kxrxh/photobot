package weed

import (
	"context"
	"fmt"
	"time"

	"csort.ru/coffeebot/internal/classification"
	"csort.ru/coffeebot/internal/database"
	"csort.ru/coffeebot/internal/dto"
	"csort.ru/coffeebot/internal/minio"
	"csort.ru/coffeebot/internal/weed/analysis"
	"csort.ru/coffeebot/internal/weed/image"
	"csort.ru/coffeebot/internal/weed/stats"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kxrxh/gopt"
	"github.com/rs/zerolog"
)

type Service struct {
	queries             database.Querier
	txRunner            database.TxRunner
	objectStore         minio.ObjectStorage
	weedStatsService    *stats.Service
	weedImageService    *image.Service
	weedAnalysesService *analysis.Service
	log                 zerolog.Logger
}

var weedQueryConverter queryConverter = &queryConverterImpl{}

// NewService wires a weed service from explicit dependencies (tests); production code uses NewServiceWithPool.
func NewService(
	queries database.Querier,
	txRunner database.TxRunner,
	objectStore minio.ObjectStorage,
	weedStatsService *stats.Service,
	weedImageService *image.Service,
	weedAnalysesService *analysis.Service,
	log zerolog.Logger,
) *Service {
	return &Service{
		queries:             queries,
		txRunner:            txRunner,
		objectStore:         objectStore,
		weedStatsService:    weedStatsService,
		weedImageService:    weedImageService,
		weedAnalysesService: weedAnalysesService,
		log:                 log,
	}
}

// NewServiceWithPool is the production constructor (pool, MinIO client, sub-services, TxRunner).
func NewServiceWithPool(
	queries *database.Queries,
	minioClient *minio.Client,
	dbPool *pgxpool.Pool,
	baseLog zerolog.Logger,
) *Service {
	weedLog := baseLog.With().Str("component", "services.weed").Logger()
	imgLog := baseLog.With().Str("component", "services.weed_images").Logger()
	txRunner := database.NewPoolTxRunner(dbPool, queries)
	weedStatsService := stats.NewService(queries, dbPool)
	weedImageService := image.NewService(queries, minioClient, imgLog)
	weedAnalysesService := analysis.NewService(queries)
	return NewService(
		queries,
		txRunner,
		minioClient,
		weedStatsService,
		weedImageService,
		weedAnalysesService,
		weedLog,
	)
}

func (s *Service) generatePresignedURLOrEmpty(ctx context.Context, key string) string {
	if s.objectStore == nil {
		return ""
	}
	url, err := s.objectStore.GeneratePresignedURL(ctx, key, time.Hour*24)
	if err != nil {
		s.log.Error().Err(err).Str("imageKey", key).Msg("failed to generate presigned url")
		return ""
	}
	return url
}

func (s *Service) deriveDimensionsFromParamsOrStats(
	params SaveWeedParams,
	fallbackLength, fallbackWidth float32,
) (float32, float32) {
	if params.Statistics != nil {
		return params.Statistics.LMedian, params.Statistics.WMedian
	}
	if fallbackLength != 0 || fallbackWidth != 0 {
		return fallbackLength, fallbackWidth
	}
	return params.Length, params.Width
}

func (s *Service) deriveDimensionsForUpdate(
	ctx context.Context,
	id int32,
	params SaveWeedParams,
) (float32, float32) {
	if params.Statistics != nil {
		return params.Statistics.LMedian, params.Statistics.WMedian
	}
	if params.Length != 0 && params.Width != 0 {
		return params.Length, params.Width
	}
	existing, err := s.queries.GetWeedByID(ctx, id)
	if err == nil {
		var l float32
		var w float32
		if params.Length == 0 && existing.Length != nil {
			l = *existing.Length
		} else {
			l = params.Length
		}
		if params.Width == 0 && existing.Width != nil {
			w = *existing.Width
		} else {
			w = params.Width
		}
		return l, w
	}
	return params.Length, params.Width
}

func (s *Service) createStatsIfProvided(
	ctx context.Context,
	qTx database.Querier,
	weedID int32,
	params *SaveWeedParams,
) error {
	if params.Statistics == nil {
		return nil
	}
	if len(params.ExcludedObjects) > 0 {
		params.Statistics.ExcludedObjects = params.ExcludedObjects
	}
	return s.weedStatsService.CreateWeedStatsWithTx(ctx, qTx, weedID, params.Statistics)
}

func (s *Service) updateStatsIfProvided(
	ctx context.Context,
	qTx database.Querier,
	weedID int32,
	params *SaveWeedParams,
) error {
	if params.Statistics == nil {
		return nil
	}
	if len(params.ExcludedObjects) > 0 {
		params.Statistics.ExcludedObjects = params.ExcludedObjects
	}
	return s.weedStatsService.UpdateWeedStatsWithTx(ctx, qTx, weedID, params.Statistics)
}

func (s *Service) replaceAnalysesWithTx(
	ctx context.Context,
	qTx database.Querier,
	weedID int32,
	analysisIDs []string,
) error {
	if len(analysisIDs) == 0 {
		return nil
	}
	if err := qTx.DeleteWeedAnalysesByWeedID(ctx, weedID); err != nil {
		return fmt.Errorf("failed to clear existing analyses: %w", err)
	}
	return s.weedAnalysesService.CreateMultipleWeedAnalysesWithTx(ctx, qTx, weedID, analysisIDs)
}

func (s *Service) ListWeeds(
	ctx context.Context,
	params ListWeedsParams,
) (*dto.PaginatedResponse[WeedListItem], error) {
	listParams := weedQueryConverter.ListWeedsParamsToListDB(params)

	items, err := s.queries.ListWeeds(ctx, listParams)
	if err != nil {
		return nil, fmt.Errorf("failed to list weeds: %w", err)
	}

	countParams := weedQueryConverter.ListWeedsParamsToCountDB(listParams)
	total, err := s.queries.CountWeeds(ctx, countParams)
	if err != nil {
		return nil, fmt.Errorf("failed to count weeds: %w", err)
	}

	keys := make([]string, len(items))
	for i, w := range items {
		if w.PrimaryImageKey != nil {
			keys[i] = *w.PrimaryImageKey
		}
	}
	var presigned []string
	if s.objectStore != nil {
		needPresign := false
		for _, k := range keys {
			if k != "" {
				needPresign = true
				break
			}
		}
		if needPresign {
			var err error
			presigned, err = s.objectStore.GeneratePresignedURLs(ctx, keys, time.Hour*24)
			if err != nil {
				s.log.Error().Err(err).Msg("batch presign failed; falling back per item")
				presigned = make([]string, len(keys))
				for i, k := range keys {
					if k != "" {
						presigned[i] = s.generatePresignedURLOrEmpty(ctx, k)
					}
				}
			}
		} else {
			presigned = make([]string, len(keys))
		}
	} else {
		presigned = make([]string, len(keys))
	}

	result := make([]WeedListItem, len(items))
	for i, w := range items {
		imageURL := ""
		if i < len(presigned) {
			imageURL = presigned[i]
		}
		result[i] = WeedListItem{
			ID:              w.ID,
			Name:            w.Name,
			PrimaryImageURL: imageURL,
			Length:          gopt.FromPtr(w.Length).UnwrapOr(float32(0)),
			Width:           gopt.FromPtr(w.Width).UnwrapOr(float32(0)),
			CreatedAt:       w.CreatedAt.Time,
			UpdatedAt:       w.UpdatedAt.Time,
			MainGroup:       classification.MapMainGroup(gopt.FromPtr(w.MainGroup).UnwrapOr("")),
			MainSubgroup: classification.MapMainSubgroup(
				gopt.FromPtr(w.MainSubgroup).UnwrapOr(""),
			),
			Subgroup:     classification.MapSubgroup(gopt.FromPtr(w.Subgroup).UnwrapOr("")),
			IsQuarantine: w.IsQuarantine,
			Harmfulness:  gopt.FromPtr(w.Harmfulness).UnwrapOr(""),
		}
	}

	return &dto.PaginatedResponse[WeedListItem]{
		Data: result, Total: total, Limit: params.Limit, Offset: params.Offset,
	}, nil
}

func (s *Service) GetWeedByID(ctx context.Context, id int32) (database.Weed, error) {
	return s.queries.GetWeedByID(ctx, id)
}

func (s *Service) CreateWeed(ctx context.Context, params SaveWeedParams) (database.Weed, error) {
	var weed database.Weed
	err := s.txRunner.Run(ctx, func(qTx database.Querier) error {
		derivedLength, derivedWidth := s.deriveDimensionsFromParamsOrStats(
			params,
			params.Length,
			params.Width,
		)

		var length, width *float32
		if derivedLength != 0 {
			length = &derivedLength
		}
		if derivedWidth != 0 {
			width = &derivedWidth
		}
		createParams := database.CreateWeedParams{
			Name:         params.Name,
			LatinName:    gopt.Cond(params.LatinName != "", params.LatinName).ToPointer(),
			Description:  gopt.Cond(params.Description != "", params.Description).ToPointer(),
			Length:       length,
			Width:        width,
			MainGroup:    gopt.Cond(params.MainGroup != "", params.MainGroup).ToPointer(),
			MainSubgroup: gopt.Cond(params.MainSubgroup != "", params.MainSubgroup).ToPointer(),
			Subgroup:     gopt.Cond(params.Subgroup != "", params.Subgroup).ToPointer(),
			IsQuarantine: params.IsQuarantine,
			Harmfulness:  gopt.Cond(params.Harmfulness != "", params.Harmfulness).ToPointer(),
		}

		var createErr error
		weed, createErr = qTx.CreateWeed(ctx, createParams)
		if createErr != nil {
			return fmt.Errorf("failed to create weed: %w", createErr)
		}

		if err := s.createStatsIfProvided(ctx, qTx, weed.ID, &params); err != nil {
			return fmt.Errorf("failed to create weed stats: %w", err)
		}

		if len(params.AnalysisIDs) > 0 {
			if err := s.weedAnalysesService.CreateMultipleWeedAnalysesWithTx(
				ctx,
				qTx,
				weed.ID,
				params.AnalysisIDs,
			); err != nil {
				return fmt.Errorf("failed to create weed analyses: %w", err)
			}
		}
		return nil
	})
	if err != nil {
		return database.Weed{}, err
	}
	return weed, nil
}

func (s *Service) UpdateWeed(
	ctx context.Context,
	id int32,
	params SaveWeedParams,
) (database.Weed, error) {
	var weed database.Weed
	err := s.txRunner.Run(ctx, func(qTx database.Querier) error {
		derivedLength, derivedWidth := s.deriveDimensionsForUpdate(ctx, id, params)

		updateParams := database.UpdateWeedParams{
			ID:           id,
			Name:         params.Name,
			LatinName:    gopt.Cond(params.LatinName != "", params.LatinName).ToPointer(),
			Description:  gopt.Cond(params.Description != "", params.Description).ToPointer(),
			Length:       &derivedLength,
			Width:        &derivedWidth,
			MainGroup:    gopt.Cond(params.MainGroup != "", params.MainGroup).ToPointer(),
			MainSubgroup: gopt.Cond(params.MainSubgroup != "", params.MainSubgroup).ToPointer(),
			Subgroup:     gopt.Cond(params.Subgroup != "", params.Subgroup).ToPointer(),
			IsQuarantine: params.IsQuarantine,
			Harmfulness:  gopt.Cond(params.Harmfulness != "", params.Harmfulness).ToPointer(),
		}

		var updateErr error
		weed, updateErr = qTx.UpdateWeed(ctx, updateParams)
		if updateErr != nil {
			return fmt.Errorf("failed to update weed: %w", updateErr)
		}

		if err := s.updateStatsIfProvided(ctx, qTx, weed.ID, &params); err != nil {
			return fmt.Errorf("failed to update weed stats: %w", err)
		}

		if len(params.AnalysisIDs) > 0 {
			if err := s.replaceAnalysesWithTx(ctx, qTx, weed.ID, params.AnalysisIDs); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return database.Weed{}, err
	}
	return weed, nil
}

func (s *Service) GetWeedWithDetails(ctx context.Context, id int32) (*WeedDetails, error) {
	weedRow, err := s.queries.GetWeedWithPrimaryImage(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get weed: %w", err)
	}

	dbImages, err := s.weedImageService.GetWeedImages(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}

	keys := make([]string, len(dbImages))
	for i, img := range dbImages {
		keys[i] = img.ImageKey
	}
	var urls []string
	if s.objectStore != nil {
		var presignErr error
		urls, presignErr = s.objectStore.GeneratePresignedURLs(ctx, keys, time.Hour*24)
		if presignErr != nil {
			s.log.Error().Err(presignErr).Msg("batch presign failed; falling back per image")
			urls = make([]string, len(keys))
			for i, k := range keys {
				if k != "" {
					urls[i] = s.generatePresignedURLOrEmpty(ctx, k)
				}
			}
		}
	} else {
		urls = make([]string, len(keys))
	}

	images := make([]image.WeedImageURL, 0, len(dbImages))
	for i, img := range dbImages {
		u := ""
		if i < len(urls) {
			u = urls[i]
		}
		images = append(
			images,
			image.WeedImageURL{ID: img.ID, WeedID: img.WeedID, URL: u, IsPrimary: img.IsPrimary},
		)
	}

	analyses, err := s.weedAnalysesService.GetWeedAnalyses(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get analyses: %w", err)
	}
	analysesIDs := make([]string, 0, len(analyses))
	for _, a := range analyses {
		if _, err := uuid.Parse(a.AnalysisID); err == nil {
			analysesIDs = append(analysesIDs, a.AnalysisID)
		}
	}

	var weedStats *stats.WeedStatistics
	statistics, err := s.queries.GetWeedStatsByWeedID(ctx, id)
	if err == nil {
		excludedObjects, err := s.weedStatsService.DeserializeExcludedObjects(
			statistics.ExcludedObjects,
		)
		if err != nil {
			excludedObjects = []int64{}
		}
		weedStats = &stats.WeedStatistics{
			WAvg:            statistics.WAvg,
			WMedian:         statistics.WMedian,
			WMin:            statistics.WMin,
			WMax:            statistics.WMax,
			LAvg:            statistics.LAvg,
			LMedian:         statistics.LMedian,
			LMin:            statistics.LMin,
			LMax:            statistics.LMax,
			SqAvg:           statistics.SqAvg,
			SqMedian:        statistics.SqMedian,
			SqMin:           statistics.SqMin,
			SqMax:           statistics.SqMax,
			RAvg:            statistics.RAvg,
			RMedian:         statistics.RMedian,
			RMin:            statistics.RMin,
			RMax:            statistics.RMax,
			GAvg:            statistics.GAvg,
			GMedian:         statistics.GMedian,
			GMin:            statistics.GMin,
			GMax:            statistics.GMax,
			BAvg:            statistics.BAvg,
			BMedian:         statistics.BMedian,
			BMin:            statistics.BMin,
			BMax:            statistics.BMax,
			HAvg:            statistics.HAvg,
			HMedian:         statistics.HMedian,
			HMin:            statistics.HMin,
			HMax:            statistics.HMax,
			SAvg:            statistics.SAvg,
			SMedian:         statistics.SMedian,
			SMin:            statistics.SMin,
			SMax:            statistics.SMax,
			VAvg:            statistics.VAvg,
			VMedian:         statistics.VMedian,
			VMin:            statistics.VMin,
			VMax:            statistics.VMax,
			LwAvg:           statistics.LwAvg,
			LwMedian:        statistics.LwMedian,
			LwMin:           statistics.LwMin,
			LwMax:           statistics.LwMax,
			BrtAvg:          statistics.BrtAvg,
			BrtMedian:       statistics.BrtMedian,
			BrtMin:          statistics.BrtMin,
			BrtMax:          statistics.BrtMax,
			SolidAvg:        statistics.SolidAvg,
			SolidMedian:     statistics.SolidMedian,
			SolidMin:        statistics.SolidMin,
			SolidMax:        statistics.SolidMax,
			SqSqcrlAvg:      statistics.SqSqcrlAvg,
			SqSqcrlMedian:   statistics.SqSqcrlMedian,
			SqSqcrlMin:      statistics.SqSqcrlMin,
			SqSqcrlMax:      statistics.SqSqcrlMax,
			ExcludedObjects: excludedObjects,
		}
	} else if err != pgx.ErrNoRows {
		return nil, fmt.Errorf("failed to get weed stats: %w", err)
	}

	return &WeedDetails{
		ID:          id,
		Name:        weedRow.Name,
		Description: gopt.FromPtr(weedRow.Description).UnwrapOr(""),
		Length:      gopt.FromPtr(weedRow.Length).UnwrapOr(float32(0)),
		Width:       gopt.FromPtr(weedRow.Width).UnwrapOr(float32(0)),
		CreatedAt:   weedRow.CreatedAt.Time,
		UpdatedAt:   weedRow.UpdatedAt.Time,
		Images:      images,
		Analyses:    analysesIDs,
		Statistics:  weedStats,
		MainGroup:   classification.MapMainGroup(gopt.FromPtr(weedRow.MainGroup).UnwrapOr("")),
		MainSubgroup: classification.MapMainSubgroup(
			gopt.FromPtr(weedRow.MainSubgroup).UnwrapOr(""),
		),
		Subgroup:     classification.MapSubgroup(gopt.FromPtr(weedRow.Subgroup).UnwrapOr("")),
		IsQuarantine: weedRow.IsQuarantine,
		Harmfulness:  gopt.FromPtr(weedRow.Harmfulness).UnwrapOr(""),
	}, nil
}

func (s *Service) DeleteWeed(ctx context.Context, id int32) error {
	return s.queries.DeleteWeed(ctx, id)
}
