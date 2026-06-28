package http

import (
	"bytes"
	"strings"

	"csort.ru/analysis-service/internal/analysis"
	"csort.ru/analysis-service/internal/api/auth"
	"csort.ru/analysis-service/internal/apierrors"
	"csort.ru/analysis-service/internal/dto"
	"csort.ru/analysis-service/internal/messenger/max"
	"csort.ru/analysis-service/internal/messenger/telegram"
	"csort.ru/analysis-service/internal/middleware"
	"csort.ru/analysis-service/internal/report"
	"csort.ru/analysis-service/internal/transport/response"

	"github.com/gofiber/fiber/v3"
	"github.com/rs/zerolog"
)

const (
	headerMessengerPlatform = "X-Messenger-Platform"
	headerBotName           = "X-Bot-Name"
)

func sanitizeFilename(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	s = strings.ReplaceAll(s, `"`, "_")
	return s
}

type ReportHandler struct {
	log             zerolog.Logger
	service         *report.Service
	analysisService *analysis.Service
	authClient      *auth.Client
	telegramClient  *telegram.Client
	maxClient       *max.Client
}

func NewReportHandler(
	log zerolog.Logger,
	service *report.Service,
	analysisService *analysis.Service,
	authClient *auth.Client,
	telegramClient *telegram.Client,
	maxClient *max.Client,
) *ReportHandler {
	return &ReportHandler{
		log:             log,
		service:         service,
		analysisService: analysisService,
		authClient:      authClient,
		telegramClient:  telegramClient,
		maxClient:       maxClient,
	}
}

func (h *ReportHandler) SendToChat(c fiber.Ctx) error {
	analysisID := c.Params("id")
	if analysisID == "" {
		h.log.Warn().Msg("send to chat rejected: missing analysis_id")
		return apierrors.BadRequest("analysisId parameter is required")
	}

	identity, ok := c.Locals(middleware.UserDataKey).(*auth.Identity)
	if !ok || identity == nil {
		h.log.Warn().Msg("send to chat rejected: auth required")
		return apierrors.Unauthorized("Authentication required")
	}

	platform := strings.ToLower(strings.TrimSpace(c.Get(headerMessengerPlatform)))
	if platform == "" {
		switch {
		case identity.TelegramID != nil && identity.MaxID == nil:
			platform = "telegram"
		case identity.MaxID != nil && identity.TelegramID == nil:
			platform = "max"
		default:
			h.log.Warn().
				Msg("send to chat rejected: X-Messenger-Platform required when user has multiple platforms")
			return apierrors.BadRequest("X-Messenger-Platform header is required (telegram or max)")
		}
	}
	if platform != "telegram" && platform != "max" {
		return apierrors.BadRequest("X-Messenger-Platform must be telegram or max")
	}
	botName := strings.TrimSpace(c.Get(headerBotName))
	if botName == "" {
		return apierrors.BadRequest("X-Bot-Name header is required")
	}

	var chatID int64
	switch platform {
	case "telegram":
		if identity.TelegramID == nil {
			return apierrors.BadRequest("user has no Telegram ID")
		}
		chatID = *identity.TelegramID
	case "max":
		if identity.MaxID == nil {
			return apierrors.BadRequest("user has no MAX ID")
		}
		chatID = *identity.MaxID
	}

	domain, err := h.analysisService.GetByID(c.Context(), analysisID)
	if err != nil {
		return err
	}
	if domain.UserID != chatID {
		h.log.Warn().
			Str("analysis_id", analysisID).
			Int64("analysis_user_id", domain.Analysis.UserID).
			Int64("identity_chat_id", chatID).
			Msg("send to chat rejected: analysis ownership mismatch")
		return apierrors.Forbidden("you do not have access to this analysis")
	}

	botToken, err := h.authClient.GetBotTokenByNameAndPlatform(c.Context(), botName, platform)
	if err != nil {
		h.log.Error().
			Err(err).
			Str("bot_name", botName).
			Str("platform", platform).
			Msg("send to chat: failed to get bot token")
		return apierrors.Internal("failed to get bot token")
	}

	var buf bytes.Buffer
	_, dlErr := h.service.DownloadFile(c.Context(), analysisID, "pdf", &buf)
	if dlErr != nil {
		h.log.Error().
			Err(dlErr).
			Str("analysis_id", analysisID).
			Msg("send to chat: failed to download report")
		if _, ok := apierrors.From(dlErr); ok {
			return dlErr
		}
		return apierrors.InternalWrap(dlErr, dlErr.Error())
	}

	filename := sanitizeFilename(analysisID) + "_report.pdf"
	doc := buf.Bytes()

	switch platform {
	case "telegram":
		err = h.telegramClient.SendDocument(c.Context(), botToken, chatID, doc, filename)
	case "max":
		err = h.maxClient.SendDocument(c.Context(), botToken, chatID, doc, filename)
	}
	if err != nil {
		h.log.Error().
			Err(err).
			Str("analysis_id", analysisID).
			Str("platform", platform).
			Msg("send to chat: failed to send document")
		return apierrors.Internal("failed to send report to chat")
	}

	h.log.Info().
		Str("analysis_id", analysisID).
		Str("platform", platform).
		Msg("report sent to chat successfully")
	return response.OK(c, dto.MessageResponse{Message: "Report sent to chat"})
}
