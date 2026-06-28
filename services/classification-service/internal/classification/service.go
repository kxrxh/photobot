package classification

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gofiber/fiber/v3"

	database "csort.ru/classification-service/internal/database"
	"csort.ru/classification-service/internal/httperr"
	"csort.ru/classification-service/internal/logger"
	"github.com/bytedance/sonic"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type ClassificationService struct {
	q      *database.Queries
	pool   *pgxpool.Pool
	logger zerolog.Logger
}

func NewClassificationService(q *database.Queries, pool *pgxpool.Pool) *ClassificationService {
	return &ClassificationService{
		q:      q,
		pool:   pool,
		logger: logger.GetLogger("services.classification"),
	}
}

func (s *ClassificationService) GetCompleteClassification(
	ctx context.Context,
	id uuid.UUID,
) (*CompleteClassification, error) {
	classificationRaw, err := s.q.GetCompleteClassificationByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httperr.Wrap(err, fiber.StatusNotFound, "Classification not found")
		}
		s.logger.Error().
			Err(err).
			Str("classification_id", id.String()).
			Msg("get complete classification failed")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to get complete classification",
		)
	}

	var classification Classification
	if err := sonic.Unmarshal(classificationRaw.Classification, &classification); err != nil {
		s.logger.Error().
			Err(err).
			Str("classification_id", id.String()).
			Msg("unmarshal classification failed")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to unmarshal classification JSON",
		)
	}

	var fractions []Fraction
	if len(classificationRaw.Fractions) == 0 {
		if err := sonic.Unmarshal([]byte("[]"), &fractions); err != nil {
			s.logger.Error().
				Err(err).
				Str("classification_id", id.String()).
				Msg("unmarshal fractions failed")
			return nil, httperr.Wrap(
				err,
				fiber.StatusInternalServerError,
				"Failed to unmarshal empty fractions array",
			)
		}
	} else {
		if err := sonic.Unmarshal(classificationRaw.Fractions, &fractions); err != nil {
			s.logger.Error().
				Err(err).
				Str("classification_id", id.String()).
				Msg("unmarshal fractions failed")
			return nil, httperr.Wrap(
				err,
				fiber.StatusInternalServerError,
				"Failed to unmarshal fractions JSON",
			)
		}
	}

	completeClassification := CompleteClassification{
		Classification: classification,
		Fractions:      fractions,
	}

	return &completeClassification, nil
}

func (s *ClassificationService) ListClassificationsForUser(
	ctx context.Context,
	userID int32,
	filters ClassificationFilters,
) ([]Classification, *Classification, error) {
	rows, err := s.q.GetClassificationsWithFiltersAndActive(
		ctx,
		filtersToActiveParams(userID, filters),
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("list classifications for user failed")
		return nil, nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to get classifications",
		)
	}

	result := make([]Classification, 0, len(rows))
	var active *Classification
	for _, row := range rows {
		classification := classificationFromFiltersActiveRow(row)
		result = append(result, classification)
		if isActive, ok := row.IsUserActive.(bool); ok && isActive {
			active = &classification
		}
	}

	return result, active, nil
}

func (s *ClassificationService) DeleteClassification(ctx context.Context, id uuid.UUID) error {
	err := s.q.DeleteClassification(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return httperr.Wrap(err, fiber.StatusNotFound, "Classification not found")
		}
		s.logger.Error().
			Err(err).
			Str("classification_id", id.String()).
			Msg("delete classification failed")
		return httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to delete classification")
	}

	s.logger.Debug().Str("classification_id", id.String()).Msg("delete classification completed")

	return nil
}

func (s *ClassificationService) UpdateClassificationComplete(
	ctx context.Context,
	id uuid.UUID,
	updated SaveCompleteClassificationRequest,
) (*Classification, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("classification_id", id.String()).
			Msg("begin transaction failed")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to start transaction",
		)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := s.q.WithTx(tx)

	classification, err := qtx.UpdateClassification(ctx, database.UpdateClassificationParams{
		ID:        id,
		Name:      updated.Name,
		IsPublic:  updated.IsPublic,
		ProductID: updated.Product.ID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httperr.Wrap(err, fiber.StatusNotFound, "Classification not found")
		}
		s.logger.Error().
			Err(err).
			Str("classification_id", id.String()).
			Msg("update classification failed")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to update classification",
		)
	}

	err = qtx.RemoveAllFractionsForClassification(ctx, id)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("classification_id", id.String()).
			Msg("remove fractions failed")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to update classification fractions",
		)
	}

	if _, err = s.bulkCreateFractionTree(
		ctx,
		qtx,
		classification.ID,
		updated.Fractions,
	); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Error().
			Err(err).
			Str("classification_id", id.String()).
			Msg("commit transaction failed")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to commit transaction",
		)
	}

	s.logger.Debug().
		Str("classification_id", id.String()).
		Str("name", classification.Name).
		Int("fractions_count", len(updated.Fractions)).
		Msg("update classification completed")

	return &Classification{
		ID:        classification.ID,
		Name:      classification.Name,
		CreatedBy: classification.CreatedBy,
		IsPublic:  classification.IsPublic,
		Product:   updated.Product,
		CreatedAt: classification.CreatedAt,
		UpdatedAt: classification.UpdatedAt,
	}, nil
}

func (s *ClassificationService) CreateCompleteClassification(
	ctx context.Context,
	req SaveCompleteClassificationRequest,
	createdBy int32,
) (*CompleteClassification, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.logger.Error().Err(err).Str("name", req.Name).Msg("begin transaction failed")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to start transaction",
		)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := s.q.WithTx(tx)

	createdClassification, err := s.createClassificationWithFractions(ctx, qtx, req, createdBy)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		s.logger.Error().Err(err).Str("name", req.Name).Msg("commit transaction failed")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to commit transaction",
		)
	}

	s.logger.Debug().
		Str("classification_id", createdClassification.Classification.ID.String()).
		Str("name", createdClassification.Classification.Name).
		Int("fractions_count", len(createdClassification.Fractions)).
		Msg("create classification completed")

	return createdClassification, nil
}

func (s *ClassificationService) UpdateClassificationPublic(
	ctx context.Context,
	id uuid.UUID,
	isPublic bool,
) error {
	err := s.q.UpdateClassificationPublic(ctx, database.UpdateClassificationPublicParams{
		ID:       id,
		IsPublic: isPublic,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return httperr.Wrap(err, fiber.StatusNotFound, "Classification not found")
		}
		s.logger.Error().
			Err(err).
			Str("classification_id", id.String()).
			Bool("is_public", isPublic).
			Msg("update classification public status failed")
		return httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to update classification public status",
		)
	}

	return nil
}

func (s *ClassificationService) createClassificationWithFractions(
	ctx context.Context,
	qtx *database.Queries,
	req SaveCompleteClassificationRequest,
	createdBy int32,
) (*CompleteClassification, error) {
	classification, err := qtx.CreateClassification(ctx, database.CreateClassificationParams{
		Name:      req.Name,
		IsPublic:  req.IsPublic,
		ProductID: req.Product.ID,
		CreatedBy: createdBy,
	})
	if err != nil {
		s.logger.Error().Err(err).Str("name", req.Name).Msg("create classification failed")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to create classification",
		)
	}

	createdFractions, err := s.bulkCreateFractionTree(ctx, qtx, classification.ID, req.Fractions)
	if err != nil {
		return nil, err
	}

	return &CompleteClassification{
		Classification: Classification{
			ID:        classification.ID,
			Name:      classification.Name,
			CreatedBy: classification.CreatedBy,
			IsPublic:  classification.IsPublic,
			Product:   req.Product,
			CreatedAt: classification.CreatedAt,
			UpdatedAt: classification.UpdatedAt,
		},
		Fractions: createdFractions,
	}, nil
}
