package params

import (
	"context"

	"csort.ru/classification-service/internal/database"
	"csort.ru/classification-service/internal/httperr"
	"csort.ru/classification-service/internal/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type ClassificationParamsService struct {
	q      *database.Queries
	logger zerolog.Logger
}

func NewClassificationParamsService(q *database.Queries) *ClassificationParamsService {
	return &ClassificationParamsService{
		q:      q,
		logger: logger.GetLogger("services.classification_params"),
	}
}

func (s *ClassificationParamsService) Create(
	ctx context.Context,
	req SaveClassificationParam,
) (*ClassificationParam, error) {
	exists, err := s.q.ClassificationParamExistsByName(ctx, req.Name)
	if err != nil {
		s.logger.Error().Err(err).Str("name", req.Name).Msg("failed to check param existence")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to check parameter existence",
		)
	}
	if exists {
		s.logger.Warn().Str("name", req.Name).Msg("classification param already exists")
		return nil, httperr.New(
			fiber.StatusConflict,
			"Classification parameter with this name already exists",
		)
	}

	created, err := s.q.CreateClassificationParam(ctx, req.Name)
	if err != nil {
		s.logger.Error().Err(err).Str("name", req.Name).Msg("failed to create classification param")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to create classification parameter",
		)
	}

	return &ClassificationParam{
		ID:        created.ID,
		Name:      created.Name,
		CreatedAt: created.CreatedAt,
		UpdatedAt: created.UpdatedAt,
	}, nil
}

func (s *ClassificationParamsService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.q.DeleteClassificationParamByID(ctx, id); err != nil {
		s.logger.Error().
			Err(err).
			Str("id", id.String()).
			Msg("failed to delete classification param")
		return httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to delete classification parameter",
		)
	}
	return nil
}

func (s *ClassificationParamsService) List(ctx context.Context) ([]ClassificationParam, error) {
	items, err := s.q.GetAllClassificationParams(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to list classification params")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to list classification parameters",
		)
	}
	result := make([]ClassificationParam, 0, len(items))
	for _, it := range items {
		result = append(result, ClassificationParam{
			ID:        it.ID,
			Name:      it.Name,
			CreatedAt: it.CreatedAt,
			UpdatedAt: it.UpdatedAt,
		})
	}
	return result, nil
}
