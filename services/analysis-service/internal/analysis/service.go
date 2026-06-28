package analysis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"csort.ru/analysis-service/internal/apierrors"
	"csort.ru/analysis-service/internal/core"
	"csort.ru/analysis-service/internal/logger"
	"csort.ru/analysis-service/internal/objects"
	"csort.ru/analysis-service/internal/observability"
	"csort.ru/analysis-service/internal/repository/kalibr"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/otel/attribute"
)

type mergeAnalysesClient interface {
	MergeAnalyses(ctx context.Context, payload map[string]any) ([]byte, error)
}

type Service struct {
	repo           *kalibr.Queries
	analysisClient mergeAnalysesClient
	logger         zerolog.Logger
}

func NewAnalysisService(repo *kalibr.Queries, analysisClient mergeAnalysesClient) *Service {
	return &Service{
		repo:           repo,
		analysisClient: analysisClient,
		logger:         logger.GetLogger("analysis.service"),
	}
}

type ListParams struct {
	Limit     int32
	Offset    int32
	Product   string
	IDFilter  string
	SortBy    string
	SortOrder string
}

func (s *Service) List(
	ctx context.Context,
	userIDs []int64,
	params ListParams,
) (*core.PaginatedResponse[AnalysisListItem], error) {
	s.logger.Debug().
		Interface("userIDs", userIDs).
		Interface("params", params).
		Msg("listing analyses")

	overallStart := time.Now()
	dbStart := time.Now()

	idExact, idPrefix := idFilterParams(params.IDFilter)

	var listRows []kalibr.GetAnalysesListRow
	var count int64
	var dbErr error

	if len(userIDs) == 1 {
		count, dbErr = s.repo.CountAnalyses(ctx, kalibr.CountAnalysesParams{
			UserID:   userIDs[0],
			Product:  params.Product,
			IDExact:  idExact,
			IDPrefix: idPrefix,
		})
		if dbErr == nil {
			listRows, dbErr = s.repo.GetAnalysesList(ctx, kalibr.GetAnalysesListParams{
				UserID:    userIDs[0],
				Product:   params.Product,
				IDExact:   idExact,
				IDPrefix:  idPrefix,
				SortBy:    params.SortBy,
				SortOrder: params.SortOrder,
				Offset:    params.Offset,
				Limit:     params.Limit,
			})
		}
	} else {
		count, dbErr = s.repo.CountAnalysesByIdUsers(ctx, kalibr.CountAnalysesByIdUsersParams{
			UserIds:  userIDs,
			Product:  params.Product,
			IDExact:  idExact,
			IDPrefix: idPrefix,
		})
		if dbErr == nil {
			byIDUsersRows, err := s.repo.GetAnalysesListByIdUsers(
				ctx,
				kalibr.GetAnalysesListByIdUsersParams{
					UserIds:   userIDs,
					Product:   params.Product,
					IDExact:   idExact,
					IDPrefix:  idPrefix,
					SortBy:    params.SortBy,
					SortOrder: params.SortOrder,
					Offset:    params.Offset,
					Limit:     params.Limit,
				},
			)
			if err != nil {
				dbErr = err
			} else {
				listRows = listRowsFromByIdUsers(byIDUsersRows)
			}
		}
	}

	if dbErr != nil {
		observability.RecordError(ctx, dbErr,
			attribute.String("analysis.operation", "list.query_db"),
			attribute.Int("analysis.user_ids_count", len(userIDs)),
		)
		s.logger.Error().
			Err(dbErr).
			Interface("userIDs", userIDs).
			Interface("params", params).
			Msg("failed to fetch analyses")
		return nil, apierrors.InternalWrap(dbErr, "failed to fetch analyses")
	}

	dbDuration := time.Since(dbStart)

	items := make([]AnalysisListItem, 0, len(listRows))

	convertStart := time.Now()
	for _, r := range listRows {
		items = append(items, convertListRowToItem(r))
	}
	convertDuration := time.Since(convertStart)

	s.logger.Debug().
		Interface("userIDs", userIDs).
		Int("rows", len(listRows)).
		Dur("db_duration", dbDuration).
		Dur("convert_duration", convertDuration).
		Dur("total_duration", time.Since(overallStart)).
		Msg("List analyses timings")

	return &core.PaginatedResponse[AnalysisListItem]{
		Data:   items,
		Total:  count,
		Limit:  params.Limit,
		Offset: params.Offset,
	}, nil
}

func (s *Service) GetByID(ctx context.Context, analysisID string) (*AnalysisWithObjects, error) {
	s.logger.Debug().Str("analysisID", analysisID).Msg("getting analysis by ID")

	id, err := uuid.Parse(analysisID)
	if err != nil {
		s.logger.Warn().Str("analysisID", analysisID).Msg("invalid analysis ID format")
		return nil, apierrors.BadRequest("invalid analysis ID")
	}

	repoAnalysis, err := s.repo.GetAnalysisByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Warn().Str("analysisID", analysisID).Msg("analysis not found")
			return nil, apierrors.NotFound("analysis not found")
		}
		observability.RecordError(ctx, err,
			attribute.String("analysis.operation", "get_by_id.load_analysis"),
			attribute.String("analysis.id", analysisID),
		)
		s.logger.Error().Err(err).Str("analysisID", analysisID).Msg("failed to get analysis")
		return nil, apierrors.InternalWrap(err, "failed to get analysis")
	}

	objs, err := objectsFromJSONB(repoAnalysis.Objects)
	if err != nil {
		observability.RecordError(ctx, err,
			attribute.String("analysis.operation", "get_by_id.parse_objects"),
			attribute.String("analysis.id", analysisID),
		)
		s.logger.Error().Err(err).Str("analysisID", analysisID).Msg("failed to parse objects JSONB")
		return nil, apierrors.InternalWrap(err, "failed to get objects")
	}

	a := convertAnalysisFromRepo(repoAnalysis)
	return &AnalysisWithObjects{
		Analysis: a,
		Objects:  objs,
	}, nil
}

func (s *Service) GetObjects(ctx context.Context, analysisID string) ([]objects.Object, error) {
	id, err := uuid.Parse(analysisID)
	if err != nil {
		s.logger.Warn().Str("analysisID", analysisID).Msg("invalid analysis ID format")
		return nil, apierrors.BadRequest("invalid analysis ID")
	}

	repoAnalysis, err := s.repo.GetAnalysisByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apierrors.NotFound("analysis not found")
		}
		observability.RecordError(ctx, err,
			attribute.String("analysis.operation", "get_objects.load_analysis"),
			attribute.String("analysis.id", analysisID),
		)
		s.logger.Error().Err(err).Str("analysisID", analysisID).Msg("failed to get analysis")
		return nil, apierrors.InternalWrap(err, "failed to get analysis")
	}

	objs, decErr := objectsFromJSONB(repoAnalysis.Objects)
	if decErr != nil {
		observability.RecordError(ctx, decErr,
			attribute.String("analysis.operation", "get_objects.decode_objects"),
			attribute.String("analysis.id", analysisID),
		)
		s.logger.Error().
			Err(decErr).
			Str("analysisID", analysisID).
			Msg("failed to decode analysis objects")
	}
	return objs, decErr
}

func objectsFromJSONB(data []byte) ([]objects.Object, error) {
	if len(data) == 0 {
		return []objects.Object{}, nil
	}
	var objs []objects.Object
	if err := json.Unmarshal(data, &objs); err != nil {
		return nil, err
	}
	for i := range objs {
		idx := int32(i)
		objs[i].ID = idx
	}
	return objs, nil
}

func (s *Service) Merge(ctx context.Context, userID int64, analysisIDs []string) error {
	payload := map[string]any{
		"user_id":  userID,
		"analyses": analysisIDs,
	}

	_, err := s.analysisClient.MergeAnalyses(ctx, payload)
	if err != nil {
		observability.RecordError(ctx, err,
			attribute.String("analysis.operation", "merge.remote"),
			attribute.Int64("analysis.merge.user_id", userID),
			attribute.Int("analysis.merge.analysis_count", len(analysisIDs)),
		)
		s.logger.Error().
			Err(err).
			Int64("user_id", userID).
			Strs("analysis_ids", analysisIDs).
			Msg("failed to merge analyses")
		return apierrors.BadGatewayWrap(err, "failed to merge analyses")
	}

	return nil
}

func listRowsFromByIdUsers(
	rows []kalibr.GetAnalysesListByIdUsersRow,
) []kalibr.GetAnalysesListRow {
	out := make([]kalibr.GetAnalysesListRow, len(rows))
	for i, r := range rows {
		out[i] = kalibr.GetAnalysesListRow(r)
	}
	return out
}

func convertListRowToItem(row kalibr.GetAnalysesListRow) AnalysisListItem {
	return AnalysisListItem{
		ID:           row.ID.String(),
		DateTime:     row.DateTime.Format(time.RFC3339),
		Product:      row.Product,
		UserID:       row.UserID,
		Source:       row.Source,
		BotMessage:   row.BotMessage,
		FilesSource:  row.FilesSource,
		FilesOutput:  row.FilesOutput,
		ScaleMmPixel: row.ScaleMmPixel,
	}
}

func convertAnalysisFromRepo(repo kalibr.Analysis) Analysis {
	return Analysis{
		ID:             repo.ID.String(),
		DateTime:       repo.DateTime.Format(time.RFC3339),
		Product:        repo.Product,
		UserID:         repo.UserID,
		Source:         repo.Source,
		BotMessage:     repo.BotMessage,
		FilesSource:    repo.FilesSource,
		FilesOutput:    repo.FilesOutput,
		ScaleMmPixel:   repo.ScaleMmPixel,
		AnalysisParams: repo.AnalysisParams,
	}
}
