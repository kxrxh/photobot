package markup

import (
	database "csort.ru/classification-service/internal/database"
	"github.com/google/uuid"
)

func groupFractionsByMarkupID(
	fractions []database.MarkupFraction,
) map[uuid.UUID][]database.MarkupFraction {
	out := make(map[uuid.UUID][]database.MarkupFraction)
	for _, f := range fractions {
		out[f.MarkupID] = append(out[f.MarkupID], f)
	}
	return out
}

func groupObjectsByFractionID(
	rows []database.MarkupFractionObject,
) map[uuid.UUID][]int64 {
	out := make(map[uuid.UUID][]int64)
	for _, row := range rows {
		out[row.MarkupFractionID] = append(out[row.MarkupFractionID], row.ObjectID)
	}
	return out
}

func groupAnalysesByMarkupID(rows []database.MarkupAnalysis) map[uuid.UUID][]int64 {
	out := make(map[uuid.UUID][]int64)
	for _, row := range rows {
		out[row.MarkupID] = append(out[row.MarkupID], row.AnalysisID)
	}
	return out
}

func buildMarkupFractions(
	fractions []database.MarkupFraction,
	objectsByFraction map[uuid.UUID][]int64,
) []MarkupFraction {
	result := make([]MarkupFraction, 0, len(fractions))
	for _, f := range fractions {
		objectIDs := objectsByFraction[f.ID]
		if objectIDs == nil {
			objectIDs = []int64{}
		}
		result = append(result, MarkupFraction{
			ID:        f.ID,
			Name:      f.Name,
			CreatedAt: f.CreatedAt,
			UpdatedAt: f.UpdatedAt,
			ObjectIDs: objectIDs,
		})
	}
	return result
}
