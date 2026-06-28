package markup

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"

	database "csort.ru/classification-service/internal/database"
	"csort.ru/classification-service/internal/httperr"
	"csort.ru/classification-service/internal/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type MarkupService struct {
	q      *database.Queries
	pool   *pgxpool.Pool
	logger zerolog.Logger
}

func NewMarkupService(q *database.Queries, pool *pgxpool.Pool) *MarkupService {
	return &MarkupService{
		q:      q,
		pool:   pool,
		logger: logger.GetLogger("services.markup"),
	}
}

func (s *MarkupService) CreateMarkup(ctx context.Context, req SaveMarkupRequest) (*Markup, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.logger.Error().Err(err).Msg("begin transaction failed")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to start transaction",
		)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()
	q := s.q.WithTx(tx)

	markup, err := q.CreateMarkup(ctx, database.CreateMarkupParams{
		Name:      req.Name,
		CreatedBy: req.CreatedBy,
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("create markup failed")
		return nil, httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to create markup")
	}

	fractions := make([]MarkupFraction, 0, len(req.Fractions))
	for _, f := range req.Fractions {
		frac, err := q.CreateMarkupFraction(ctx, database.CreateMarkupFractionParams{
			MarkupID: markup.ID,
			Name:     f.Name,
		})
		if err != nil {
			s.logger.Error().Err(err).Msg("create markup fraction failed")
			return nil, httperr.Wrap(
				err,
				fiber.StatusInternalServerError,
				"Failed to create markup fraction",
			)
		}
		if len(f.ObjectIDs) > 0 {
			err = q.BulkAddObjectsToMarkupFraction(
				ctx,
				database.BulkAddObjectsToMarkupFractionParams{
					MarkupFractionID: frac.ID,
					Column2:          f.ObjectIDs,
				},
			)
			if err != nil {
				s.logger.Error().Err(err).Msg("add objects to fraction failed")
				return nil, httperr.Wrap(
					err,
					fiber.StatusInternalServerError,
					"Failed to add object to fraction",
				)
			}
		}
		fractions = append(fractions, MarkupFraction{
			ID:        frac.ID,
			Name:      frac.Name,
			CreatedAt: frac.CreatedAt,
			UpdatedAt: frac.UpdatedAt,
			ObjectIDs: f.ObjectIDs,
		})
	}

	if len(req.AnalysesIDs) > 0 {
		err = q.BulkCreateMarkupAnalyses(ctx, database.BulkCreateMarkupAnalysesParams{
			MarkupID:    markup.ID,
			AnalysisIds: req.AnalysesIDs,
		})
		if err != nil {
			s.logger.Error().Err(err).Msg("link analysis to markup failed")
			return nil, httperr.Wrap(
				err,
				fiber.StatusInternalServerError,
				"Failed to link analysis to markup",
			)
		}
	}

	return &Markup{
		ID:          markup.ID,
		Name:        markup.Name,
		CreatedBy:   markup.CreatedBy,
		CreatedAt:   markup.CreatedAt,
		UpdatedAt:   markup.UpdatedAt,
		Fractions:   fractions,
		AnalysesIDs: req.AnalysesIDs,
	}, nil
}

func (s *MarkupService) GetMarkup(ctx context.Context, id uuid.UUID) (*Markup, error) {
	markup, err := s.q.GetMarkupByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httperr.Wrap(err, fiber.StatusNotFound, "Markup not found")
		}
		s.logger.Error().Err(err).Msg("get markup failed")
		return nil, httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to get markup")
	}

	result, err := s.loadMarkupDetails(ctx, s.q, markup)
	if err != nil {
		return nil, httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to get markup")
	}
	return result, nil
}

func (s *MarkupService) UpdateMarkup(
	ctx context.Context,
	id uuid.UUID,
	req SaveMarkupRequest,
) (*Markup, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.logger.Error().Err(err).Msg("begin transaction failed")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to start transaction",
		)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()
	q := s.q.WithTx(tx)

	_, err = q.UpdateMarkup(ctx, database.UpdateMarkupParams{
		Name: req.Name,
		ID:   id,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, httperr.Wrap(err, fiber.StatusNotFound, "Markup not found")
		}
		s.logger.Error().Err(err).Msg("update markup failed")
		return nil, httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to update markup")
	}

	currentFractions, err := q.GetMarkupFractionsByMarkupID(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Msg("get current fractions failed")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to get current fractions",
		)
	}
	for _, f := range currentFractions {
		err = q.ClearMarkupFractionObjects(ctx, f.ID)
		if err != nil {
			s.logger.Error().Err(err).Msg("clear objects for fraction failed")
			return nil, httperr.Wrap(
				err,
				fiber.StatusInternalServerError,
				"Failed to clear objects for fraction",
			)
		}
		err = q.DeleteMarkupFraction(ctx, f.ID)
		if err != nil {
			s.logger.Error().Err(err).Msg("delete fraction failed")
			return nil, httperr.Wrap(
				err,
				fiber.StatusInternalServerError,
				"Failed to delete fraction",
			)
		}
	}

	for _, f := range req.Fractions {
		frac, err := q.CreateMarkupFraction(ctx, database.CreateMarkupFractionParams{
			MarkupID: id,
			Name:     f.Name,
		})
		if err != nil {
			s.logger.Error().Err(err).Msg("create markup fraction failed")
			return nil, httperr.Wrap(
				err,
				fiber.StatusInternalServerError,
				"Failed to create markup fraction",
			)
		}
		if len(f.ObjectIDs) > 0 {
			err = q.BulkAddObjectsToMarkupFraction(
				ctx,
				database.BulkAddObjectsToMarkupFractionParams{
					MarkupFractionID: frac.ID,
					Column2:          f.ObjectIDs,
				},
			)
			if err != nil {
				s.logger.Error().Err(err).Msg("add object to fraction failed")
				return nil, httperr.Wrap(
					err,
					fiber.StatusInternalServerError,
					"Failed to add object to new fraction",
				)
			}
		}
	}

	err = q.DeleteMarkupAnalyses(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Msg("clear markup analyses failed")
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to clear markup analyses",
		)
	}
	if len(req.AnalysesIDs) > 0 {
		err = q.BulkCreateMarkupAnalyses(ctx, database.BulkCreateMarkupAnalysesParams{
			MarkupID:    id,
			AnalysisIds: req.AnalysesIDs,
		})
		if err != nil {
			s.logger.Error().Err(err).Msg("add markup analysis failed")
			return nil, httperr.Wrap(
				err,
				fiber.StatusInternalServerError,
				"Failed to add markup analysis",
			)
		}
	}

	return s.GetMarkup(ctx, id)
}

func (s *MarkupService) DeleteMarkup(ctx context.Context, id uuid.UUID) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.logger.Error().Err(err).Msg("begin transaction failed")
		return httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to start transaction")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	}()
	q := s.q.WithTx(tx)

	fractions, err := q.GetMarkupFractionsByMarkupID(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Msg("get fractions for delete failed")
		return httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to get fractions")
	}
	for _, f := range fractions {
		err = q.ClearMarkupFractionObjects(ctx, f.ID)
		if err != nil {
			s.logger.Error().Err(err).Msg("clear objects for fraction failed")
			return httperr.Wrap(
				err,
				fiber.StatusInternalServerError,
				"Failed to clear objects for fraction",
			)
		}
		err = q.DeleteMarkupFraction(ctx, f.ID)
		if err != nil {
			s.logger.Error().Err(err).Msg("delete fraction failed")
			return httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to delete fraction")
		}
	}
	err = q.DeleteMarkupAnalyses(ctx, id)
	if err != nil {
		s.logger.Error().Err(err).Msg("delete markup analyses failed")
		return httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to delete markup analyses",
		)
	}
	err = q.DeleteMarkup(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return httperr.Wrap(err, fiber.StatusNotFound, "Markup not found")
		}
		s.logger.Error().Err(err).Msg("delete markup failed")
		return httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to delete markup")
	}
	return nil
}

func (s *MarkupService) ListMarkups(ctx context.Context, filters MarkupFilters) ([]Markup, error) {
	markups, err := s.q.GetMarkups(ctx, database.GetMarkupsParams{
		CreatedBy: filters.CreatedBy,
		Name:      filters.Name,
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("list markups failed")
		return nil, httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to list markups")
	}

	result, err := s.loadMarkupsDetails(ctx, s.q, markups)
	if err != nil {
		return nil, httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to list markups")
	}
	return result, nil
}
