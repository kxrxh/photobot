package user

import (
	"context"
	"database/sql"
	"errors"

	"csort.ru/classification-service/internal/classification"
	database "csort.ru/classification-service/internal/database"
	"csort.ru/classification-service/internal/httperr"
	"csort.ru/classification-service/internal/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type UserActiveClassificationService struct {
	q               *database.Queries
	classifications *classification.ClassificationService
	logger          zerolog.Logger
}

func NewUserActiveClassificationService(
	q *database.Queries,
	classifications *classification.ClassificationService,
) *UserActiveClassificationService {
	return &UserActiveClassificationService{
		q:               q,
		classifications: classifications,
		logger:          logger.GetLogger("services.user"),
	}
}

func (s *UserActiveClassificationService) SetUserActiveClassification(
	ctx context.Context,
	req SetUserClassificationRequest,
) error {
	_, err := s.q.GetClassificationByID(ctx, req.ClassificationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return httperr.Wrap(err, fiber.StatusNotFound, "Classification not found")
		}
		s.logger.Error().
			Err(err).
			Str("classification_id", req.ClassificationID.String()).
			Msg("Classification not found")
		return httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to get classification")
	}

	_, err = s.q.SetUserActiveClassification(ctx, database.SetUserActiveClassificationParams{
		UserID:           req.UserID,
		ClassificationID: req.ClassificationID,
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to set user active classification")
		return httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to set user active classification",
		)
	}

	return nil
}

func (s *UserActiveClassificationService) GetUserActiveClassification(
	ctx context.Context,
	userID int32,
) (*classification.CompleteClassification, error) {
	userActiveClassification, err := s.q.GetUserActiveClassification(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to get user active classification",
		)
	}

	classification, err := s.classifications.GetCompleteClassification(
		ctx,
		userActiveClassification.ClassificationID,
	)
	if err != nil {
		return nil, err
	}

	return classification, nil
}

func (s *UserActiveClassificationService) DeleteUserActiveClassification(
	ctx context.Context,
	userID int32,
) error {
	err := s.q.DeleteUserActiveClassification(ctx, userID)
	if err != nil {
		s.logger.Error().
			Err(err).
			Int32("user_id", userID).
			Msg("Failed to delete user active classification")
		return httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to delete user active classification",
		)
	}

	return nil
}
