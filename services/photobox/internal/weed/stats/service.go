package stats

import (
	"context"
	"fmt"
	"strings"

	"csort.ru/coffeebot/internal/database"
	"github.com/bytedance/sonic"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	queries database.Querier
	dbPool  *pgxpool.Pool
}

func NewService(queries database.Querier, dbPool *pgxpool.Pool) *Service {
	return &Service{queries: queries, dbPool: dbPool}
}

func (s *Service) serializeExcludedObjects(excludedObjects []int64) ([]byte, error) {
	if len(excludedObjects) == 0 {
		return []byte("[]"), nil
	}
	return sonic.Marshal(excludedObjects)
}

func (s *Service) DeserializeExcludedObjects(data []byte) ([]int64, error) {
	if len(data) == 0 {
		return []int64{}, nil
	}
	var excludedObjects []int64
	if err := sonic.Unmarshal(data, &excludedObjects); err != nil {
		return []int64{}, err
	}
	return excludedObjects, nil
}

func (s *Service) CreateWeedStatsWithTx(
	ctx context.Context,
	q database.Querier,
	weedID int32,
	stats *WeedStatistics,
) error {
	excludedObjectsBytes, err := s.serializeExcludedObjects(stats.ExcludedObjects)
	if err != nil {
		return fmt.Errorf("failed to serialize excluded objects: %w", err)
	}

	params := database.CreateWeedStatsParams{
		WeedID:          weedID,
		WAvg:            stats.WAvg,
		WMedian:         stats.WMedian,
		WMin:            stats.WMin,
		WMax:            stats.WMax,
		LAvg:            stats.LAvg,
		LMedian:         stats.LMedian,
		LMin:            stats.LMin,
		LMax:            stats.LMax,
		SqAvg:           stats.SqAvg,
		SqMedian:        stats.SqMedian,
		SqMin:           stats.SqMin,
		SqMax:           stats.SqMax,
		RAvg:            stats.RAvg,
		RMedian:         stats.RMedian,
		RMin:            stats.RMin,
		RMax:            stats.RMax,
		GAvg:            stats.GAvg,
		GMedian:         stats.GMedian,
		GMin:            stats.GMin,
		GMax:            stats.GMax,
		BAvg:            stats.BAvg,
		BMedian:         stats.BMedian,
		BMin:            stats.BMin,
		BMax:            stats.BMax,
		HAvg:            stats.HAvg,
		HMedian:         stats.HMedian,
		HMin:            stats.HMin,
		HMax:            stats.HMax,
		SAvg:            stats.SAvg,
		SMedian:         stats.SMedian,
		SMin:            stats.SMin,
		SMax:            stats.SMax,
		VAvg:            stats.VAvg,
		VMedian:         stats.VMedian,
		VMin:            stats.VMin,
		VMax:            stats.VMax,
		LwAvg:           stats.LwAvg,
		LwMedian:        stats.LwMedian,
		LwMin:           stats.LwMin,
		LwMax:           stats.LwMax,
		BrtAvg:          stats.BrtAvg,
		BrtMedian:       stats.BrtMedian,
		BrtMin:          stats.BrtMin,
		BrtMax:          stats.BrtMax,
		SolidAvg:        stats.SolidAvg,
		SolidMedian:     stats.SolidMedian,
		SolidMin:        stats.SolidMin,
		SolidMax:        stats.SolidMax,
		SqSqcrlAvg:      stats.SqSqcrlAvg,
		SqSqcrlMedian:   stats.SqSqcrlMedian,
		SqSqcrlMin:      stats.SqSqcrlMin,
		SqSqcrlMax:      stats.SqSqcrlMax,
		ExcludedObjects: excludedObjectsBytes,
	}

	if _, err := q.CreateWeedStats(ctx, params); err != nil {
		return fmt.Errorf("failed to create weed stats: %w", err)
	}
	return nil
}

func (s *Service) UpdateWeedStatsWithTx(
	ctx context.Context,
	q database.Querier,
	weedID int32,
	stats *WeedStatistics,
) error {
	_, err := q.GetWeedStatsByWeedID(ctx, weedID)
	if err != nil {
		if err == pgx.ErrNoRows || strings.Contains(err.Error(), "SQLSTATE 42P01") {
			return s.CreateWeedStatsWithTx(ctx, q, weedID, stats)
		}
		return fmt.Errorf("failed to get weed stats during update: %w", err)
	}

	excludedObjectsBytes, err := s.serializeExcludedObjects(stats.ExcludedObjects)
	if err != nil {
		return fmt.Errorf("failed to serialize excluded objects: %w", err)
	}

	params := database.UpdateWeedStatsParams{
		WAvg:            stats.WAvg,
		WMedian:         stats.WMedian,
		WMin:            stats.WMin,
		WMax:            stats.WMax,
		LAvg:            stats.LAvg,
		LMedian:         stats.LMedian,
		LMin:            stats.LMin,
		LMax:            stats.LMax,
		SqAvg:           stats.SqAvg,
		SqMedian:        stats.SqMedian,
		SqMin:           stats.SqMin,
		SqMax:           stats.SqMax,
		RAvg:            stats.RAvg,
		RMedian:         stats.RMedian,
		RMin:            stats.RMin,
		RMax:            stats.RMax,
		GAvg:            stats.GAvg,
		GMedian:         stats.GMedian,
		GMin:            stats.GMin,
		GMax:            stats.GMax,
		BAvg:            stats.BAvg,
		BMedian:         stats.BMedian,
		BMin:            stats.BMin,
		BMax:            stats.BMax,
		HAvg:            stats.HAvg,
		HMedian:         stats.HMedian,
		HMin:            stats.HMin,
		HMax:            stats.HMax,
		SAvg:            stats.SAvg,
		SMedian:         stats.SMedian,
		SMin:            stats.SMin,
		SMax:            stats.SMax,
		VAvg:            stats.VAvg,
		VMedian:         stats.VMedian,
		VMin:            stats.VMin,
		VMax:            stats.VMax,
		LwAvg:           stats.LwAvg,
		LwMedian:        stats.LwMedian,
		LwMin:           stats.LwMin,
		LwMax:           stats.LwMax,
		BrtAvg:          stats.BrtAvg,
		BrtMedian:       stats.BrtMedian,
		BrtMin:          stats.BrtMin,
		BrtMax:          stats.BrtMax,
		SolidAvg:        stats.SolidAvg,
		SolidMedian:     stats.SolidMedian,
		SolidMin:        stats.SolidMin,
		SolidMax:        stats.SolidMax,
		SqSqcrlAvg:      stats.SqSqcrlAvg,
		SqSqcrlMedian:   stats.SqSqcrlMedian,
		SqSqcrlMin:      stats.SqSqcrlMin,
		SqSqcrlMax:      stats.SqSqcrlMax,
		ExcludedObjects: excludedObjectsBytes,
		WeedID:          weedID,
	}

	if _, err := q.UpdateWeedStats(ctx, params); err != nil {
		return fmt.Errorf("failed to update weed stats during update: %w", err)
	}
	return nil
}
