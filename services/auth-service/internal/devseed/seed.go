package devseed

import (
	"context"

	"csort.ru/auth-service/internal/bot"
	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/logger"
	"csort.ru/auth-service/internal/webauth"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	DevLogin      = "dev"
	DevPassword   = "dev"
	DevBotName    = "photobot"
	DevBotToken   = "dev-bot-token" //nolint:gosec // DEV_MODE placeholder; not a production secret
	DevTelegramID = int64(919216442)
)

func Run(ctx context.Context, db *database.Queries, botSvc *bot.Service) error {
	log := logger.GetLogger("devseed")

	hash, err := webauth.HashPassword(DevPassword)
	if err != nil {
		return err
	}

	user, err := db.UpsertDevUser(ctx, database.UpsertDevUserParams{
		Login:        pgtype.Text{String: DevLogin, Valid: true},
		PasswordHash: pgtype.Text{String: hash, Valid: true},
		TelegramID:   pgtype.Int8{Int64: DevTelegramID, Valid: true},
	})
	if err != nil {
		return err
	}
	log.Info().Int32("user_id", user.ID).Str("login", DevLogin).Msg("dev user ready")

	if botSvc == nil {
		return nil
	}

	if _, err := db.GetBotByNameAndPlatform(ctx, database.GetBotByNameAndPlatformParams{
		Name: DevBotName, Platform: "telegram",
	}); err == nil {
		log.Info().Str("bot", DevBotName).Msg("dev bot already exists")
		return nil
	}

	_, err = botSvc.Create(ctx, bot.CreateRequest{
		Name:     DevBotName,
		Token:    DevBotToken,
		Platform: "telegram",
	})
	if err != nil {
		log.Warn().Err(err).Msg("dev bot seed skipped (may already exist)")
		return nil
	}
	log.Info().Str("bot", DevBotName).Msg("dev bot created")
	return nil
}
