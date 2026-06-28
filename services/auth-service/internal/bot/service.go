package bot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"csort.ru/auth-service/internal/apperrors"
	"csort.ru/auth-service/internal/crypto"
	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/logger"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

type Service struct {
	db            *database.Queries
	encryptionKey []byte
	logger        zerolog.Logger
}

// NewService returns a bot service; encryptionKey must be 32 bytes.
func NewService(db *database.Queries, encryptionKey string) (*Service, error) {
	keyBytes := []byte(encryptionKey)
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf(
			"encryption key must be 32 bytes long, got %d bytes (string length: %d)",
			len(keyBytes),
			len(encryptionKey),
		)
	}
	return &Service{
		db:            db,
		encryptionKey: keyBytes,
		logger:        logger.GetLogger("bot.service"),
	}, nil
}

const defaultPlatform = "telegram"

func (s *Service) Create(ctx context.Context, req CreateRequest) (*Response, error) {
	platform := req.Platform
	if platform == "" {
		platform = defaultPlatform
	}

	encryptedToken, err := crypto.Encrypt([]byte(req.Token), s.encryptionKey)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to encrypt token")
		return nil, fmt.Errorf("could not create bot: %w", err)
	}

	bot, err := s.db.CreateBot(ctx, database.CreateBotParams{
		Name:     req.Name,
		Token:    encryptedToken,
		Platform: platform,
	})
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to create bot in db")
		return nil, err
	}
	return &Response{
		ID:        bot.ID,
		Name:      bot.Name,
		Platform:  bot.Platform,
		CreatedAt: bot.CreatedAt,
		UpdatedAt: bot.UpdatedAt,
	}, nil
}

func (s *Service) GetByName(ctx context.Context, name string) (*Bot, error) {
	bot, err := s.db.GetBotByName(ctx, name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.New(fiber.StatusNotFound, "bot not found")
		}
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to get bot from database",
		)
	}
	decryptedToken, err := crypto.Decrypt(bot.Token, s.encryptionKey)
	if err != nil {
		s.logger.Error().Err(err).Str("bot_name", name).Msg("Failed to decrypt token for bot")
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to decrypt bot token",
		)
	}
	return &Bot{
		ID:        bot.ID,
		Name:      bot.Name,
		Platform:  bot.Platform,
		Token:     string(decryptedToken),
		CreatedAt: bot.CreatedAt,
		UpdatedAt: bot.UpdatedAt,
	}, nil
}

func (s *Service) List(ctx context.Context) ([]Response, error) {
	bots, err := s.db.ListBots(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to list bots")
	}
	res := make([]Response, len(bots))
	for i, bot := range bots {
		res[i] = Response{
			ID:        bot.ID,
			Name:      bot.Name,
			Platform:  bot.Platform,
			CreatedAt: bot.CreatedAt,
			UpdatedAt: bot.UpdatedAt,
		}
	}
	return res, nil
}

func (s *Service) Update(ctx context.Context, id int32, req UpdateRequest) (*Response, error) {
	var updatedBot *Response

	if req.Name != nil {
		bot, err := s.db.UpdateBotName(ctx, database.UpdateBotNameParams{
			ID:   id,
			Name: *req.Name,
		})
		if err != nil {
			return nil, apperrors.Wrap(
				err,
				fiber.StatusInternalServerError,
				"failed to update bot name",
			)
		}
		updatedBot = &Response{
			ID:        bot.ID,
			Name:      bot.Name,
			Platform:  bot.Platform,
			CreatedAt: bot.CreatedAt,
			UpdatedAt: bot.UpdatedAt,
		}
	}

	if req.Token != nil {
		encryptedToken, err := crypto.Encrypt([]byte(*req.Token), s.encryptionKey)
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to encrypt new token for update")
			return nil, apperrors.Wrap(
				err,
				fiber.StatusInternalServerError,
				"failed to encrypt bot token",
			)
		}

		bot, err := s.db.UpdateBotToken(ctx, database.UpdateBotTokenParams{
			ID:    id,
			Token: encryptedToken,
		})
		if err != nil {
			return nil, apperrors.Wrap(
				err,
				fiber.StatusInternalServerError,
				"failed to update bot token",
			)
		}
		updatedBot = &Response{
			ID:        bot.ID,
			Name:      bot.Name,
			Platform:  bot.Platform,
			CreatedAt: bot.CreatedAt,
			UpdatedAt: bot.UpdatedAt,
		}
	}

	if updatedBot == nil {
		return nil, apperrors.New(fiber.StatusBadRequest, "No fields to update provided")
	}

	return updatedBot, nil
}

func (s *Service) Delete(ctx context.Context, id int32) error {
	if err := s.db.DeleteBot(ctx, id); err != nil {
		return apperrors.Wrap(err, fiber.StatusInternalServerError, "failed to delete bot")
	}
	return nil
}

// GetTokenByNameAndPlatform returns decrypted token for the given bot name and platform.
func (s *Service) GetTokenByNameAndPlatform(
	ctx context.Context,
	name, platform string,
) (string, error) {
	if platform == "" {
		platform = "telegram"
	}

	bot, err := s.db.GetBotByNameAndPlatform(ctx, database.GetBotByNameAndPlatformParams{
		Name:     name,
		Platform: platform,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", apperrors.New(fiber.StatusNotFound, "bot not found")
		}
		return "", apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to get bot from database",
		)
	}
	decryptedToken, err := crypto.Decrypt(bot.Token, s.encryptionKey)
	if err != nil {
		s.logger.Error().
			Err(err).
			Str("bot_name", name).
			Str("platform", platform).
			Msg("Failed to decrypt token for bot")
		return "", apperrors.New(fiber.StatusInternalServerError, "failed to decrypt bot token")
	}
	return string(decryptedToken), nil
}

func (s *Service) ListTokensByPlatform(ctx context.Context, platform string) ([]string, error) {
	if platform == "" {
		platform = defaultPlatform
	}

	bots, err := s.db.ListBotsByPlatform(ctx, platform)
	if err != nil {
		return nil, apperrors.Wrap(
			err,
			fiber.StatusInternalServerError,
			"failed to list bots by platform",
		)
	}

	tokens := make([]string, 0, len(bots))
	for _, bot := range bots {
		decryptedToken, decErr := crypto.Decrypt(bot.Token, s.encryptionKey)
		if decErr != nil {
			s.logger.Warn().
				Err(decErr).
				Str("bot_name", bot.Name).
				Str("platform", platform).
				Msg("Failed to decrypt bot token; skipping")
			continue
		}
		tokens = append(tokens, string(decryptedToken))
	}
	return tokens, nil
}
