package ownership

import (
	"context"

	database "csort.ru/classification-service/internal/database"
	"csort.ru/classification-service/internal/httperr"
	"csort.ru/classification-service/internal/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

type Service struct {
	q      *database.Queries
	pool   *pgxpool.Pool
	logger zerolog.Logger
}

func NewService(q *database.Queries, pool *pgxpool.Pool) *Service {
	return &Service{
		q:      q,
		pool:   pool,
		logger: logger.GetLogger("services.merge"),
	}
}

func (s *Service) TransferOwnership(ctx context.Context, fromUserID, toUserID int32) (err error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.logger.Error().Err(err).Msg("begin transaction failed")
		return httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to transfer ownership")
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	q := s.q.WithTx(tx)

	err = q.MergeReassignClassifications(ctx, database.MergeReassignClassificationsParams{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
	})
	if err != nil {
		s.logger.Error().
			Err(err).
			Int32("from_user_id", fromUserID).
			Int32("to_user_id", toUserID).
			Msg("reassign classifications failed")
		return httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to transfer ownership")
	}

	err = q.MergeReassignMarkups(ctx, database.MergeReassignMarkupsParams{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
	})
	if err != nil {
		s.logger.Error().
			Err(err).
			Int32("from_user_id", fromUserID).
			Int32("to_user_id", toUserID).
			Msg("reassign markups failed")
		return httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to transfer ownership")
	}

	rows, err := q.MergeGetUserActiveClassifications(
		ctx,
		database.MergeGetUserActiveClassificationsParams{
			FromUserID: fromUserID,
			ToUserID:   toUserID,
		},
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("get user active classifications for merge failed")
		return httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to transfer ownership")
	}

	var fromRow, toRow *database.MergeGetUserActiveClassificationsRow
	for i := range rows {
		if rows[i].UserID == fromUserID {
			fromRow = &rows[i]
		} else {
			toRow = &rows[i]
		}
	}

	if fromRow != nil && toRow != nil {
		if !toRow.UpdatedAt.After(fromRow.UpdatedAt) {
			if err = q.MergeUserActiveClassificationDeleteByUserID(ctx, toUserID); err != nil {
				return httperr.Wrap(
					err,
					fiber.StatusInternalServerError,
					"Failed to transfer ownership",
				)
			}
			if err = q.MergeUserActiveClassificationReassign(
				ctx,
				database.MergeUserActiveClassificationReassignParams{
					FromUserID: fromUserID,
					ToUserID:   toUserID,
				},
			); err != nil {
				return httperr.Wrap(
					err,
					fiber.StatusInternalServerError,
					"Failed to transfer ownership",
				)
			}
		} else {
			if err = q.MergeUserActiveClassificationDeleteByUserID(ctx, fromUserID); err != nil {
				return httperr.Wrap(
					err,
					fiber.StatusInternalServerError,
					"Failed to transfer ownership",
				)
			}
		}
	} else if fromRow != nil {
		if err = q.MergeUserActiveClassificationReassign(
			ctx,
			database.MergeUserActiveClassificationReassignParams{
				FromUserID: fromUserID,
				ToUserID:   toUserID,
			},
		); err != nil {
			return httperr.Wrap(
				err,
				fiber.StatusInternalServerError,
				"Failed to transfer ownership",
			)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		s.logger.Error().Err(err).Msg("commit merge transaction failed")
		return httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to transfer ownership")
	}

	s.logger.Info().
		Int32("from_user_id", fromUserID).
		Int32("to_user_id", toUserID).
		Msg("ownership transfer completed")
	return nil
}
