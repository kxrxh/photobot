package analysis

import (
	"context"

	"csort.ru/coffeebot/internal/database"
)

type Service struct {
	queries database.Querier
}

func NewService(queries database.Querier) *Service {
	return &Service{queries: queries}
}

func (s *Service) GetWeedAnalyses(
	ctx context.Context,
	weedID int32,
) ([]database.WeedAnalysis, error) {
	return s.queries.GetWeedAnalyses(ctx, weedID)
}

func (s *Service) CreateMultipleWeedAnalysesWithTx(
	ctx context.Context,
	q database.Querier,
	weedID int32,
	analysisIDs []string,
) error {
	if len(analysisIDs) == 0 {
		return nil
	}
	return q.BulkInsertWeedAnalyses(ctx, database.BulkInsertWeedAnalysesParams{
		WeedID:      weedID,
		AnalysisIds: analysisIDs,
	})
}
