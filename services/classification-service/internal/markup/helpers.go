package markup

import (
	"context"

	database "csort.ru/classification-service/internal/database"
	"github.com/google/uuid"
)

func (s *MarkupService) loadMarkupDetails(
	ctx context.Context,
	q *database.Queries,
	markup database.Markup,
) (*Markup, error) {
	fractionsDB, err := q.GetMarkupFractionsByMarkupID(ctx, markup.ID)
	if err != nil {
		s.logger.Error().Err(err).Msg("get markup fractions failed")
		return nil, err
	}

	fractionIDs := make([]uuid.UUID, 0, len(fractionsDB))
	for _, f := range fractionsDB {
		fractionIDs = append(fractionIDs, f.ID)
	}

	var objectsByFraction map[uuid.UUID][]int64
	if len(fractionIDs) > 0 {
		objectRows, err := q.GetObjectsByMarkupFractionIDs(ctx, fractionIDs)
		if err != nil {
			s.logger.Error().Err(err).Msg("get objects for fractions failed")
			return nil, err
		}
		objectsByFraction = groupObjectsByFractionID(objectRows)
	} else {
		objectsByFraction = map[uuid.UUID][]int64{}
	}

	analysesIDs, err := q.GetMarkupAnalysesByMarkupID(ctx, markup.ID)
	if err != nil {
		s.logger.Error().Err(err).Msg("get analyses for markup failed")
		return nil, err
	}

	return &Markup{
		ID:          markup.ID,
		Name:        markup.Name,
		CreatedBy:   markup.CreatedBy,
		CreatedAt:   markup.CreatedAt,
		UpdatedAt:   markup.UpdatedAt,
		Fractions:   buildMarkupFractions(fractionsDB, objectsByFraction),
		AnalysesIDs: analysesIDs,
	}, nil
}

func (s *MarkupService) loadMarkupsDetails(
	ctx context.Context,
	q *database.Queries,
	markups []database.Markup,
) ([]Markup, error) {
	if len(markups) == 0 {
		return []Markup{}, nil
	}

	markupIDs := make([]uuid.UUID, 0, len(markups))
	for _, m := range markups {
		markupIDs = append(markupIDs, m.ID)
	}

	fractionsDB, err := q.GetMarkupFractionsByMarkupIDs(ctx, markupIDs)
	if err != nil {
		s.logger.Error().Err(err).Msg("get markup fractions failed")
		return nil, err
	}

	fractionIDs := make([]uuid.UUID, 0, len(fractionsDB))
	for _, f := range fractionsDB {
		fractionIDs = append(fractionIDs, f.ID)
	}

	var objectsByFraction map[uuid.UUID][]int64
	if len(fractionIDs) > 0 {
		objectRows, err := q.GetObjectsByMarkupFractionIDs(ctx, fractionIDs)
		if err != nil {
			s.logger.Error().Err(err).Msg("get objects for fractions failed")
			return nil, err
		}
		objectsByFraction = groupObjectsByFractionID(objectRows)
	} else {
		objectsByFraction = map[uuid.UUID][]int64{}
	}

	analysisRows, err := q.GetMarkupAnalysesByMarkupIDs(ctx, markupIDs)
	if err != nil {
		s.logger.Error().Err(err).Msg("get analyses for markups failed")
		return nil, err
	}

	fractionsByMarkup := groupFractionsByMarkupID(fractionsDB)
	analysesByMarkup := groupAnalysesByMarkupID(analysisRows)

	result := make([]Markup, 0, len(markups))
	for _, m := range markups {
		analysesIDs := analysesByMarkup[m.ID]
		if analysesIDs == nil {
			analysesIDs = []int64{}
		}
		result = append(result, Markup{
			ID:          m.ID,
			Name:        m.Name,
			CreatedBy:   m.CreatedBy,
			CreatedAt:   m.CreatedAt,
			UpdatedAt:   m.UpdatedAt,
			Fractions:   buildMarkupFractions(fractionsByMarkup[m.ID], objectsByFraction),
			AnalysesIDs: analysesIDs,
		})
	}
	return result, nil
}
