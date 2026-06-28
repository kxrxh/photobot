package http

import (
	"fmt"
	"reflect"
	"strconv"

	"csort.ru/auth-service/internal/apperrors"
	"csort.ru/auth-service/internal/bot"
	"csort.ru/auth-service/internal/dto"
	"csort.ru/auth-service/internal/transport/response"
	"csort.ru/auth-service/internal/validation"
	validatepkg "csort.ru/auth-service/pkg/validator"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type BotHandler struct {
	botService *bot.Service
	validator  *validator.Validate
}

func NewBotHandler(botService *bot.Service, v *validator.Validate) *BotHandler {
	return &BotHandler{
		botService: botService,
		validator:  v,
	}
}

// CreateBot creates a new bot
func (h *BotHandler) CreateBot(c fiber.Ctx) error {
	var req dto.CreateBotRequest

	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		var validationErrors []dto.ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, dto.ValidationError{
				Field:   validation.GetJSONFieldName(err.Field(), reflect.TypeOf(req)),
				Message: validatepkg.Translate(err),
			})
		}
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", validationErrors)
	}

	domainReq := BotCreateRequestToDomain(req)
	bot, err := h.botService.Create(c.Context(), domainReq)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return response.Created(c, BotResponseFromDomain(bot))
}

// ListBots retrieves all bots
func (h *BotHandler) ListBots(c fiber.Ctx) error {
	bots, err := h.botService.List(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return response.OK(c, BotResponsesFromDomain(bots))
}

// GetBotTokenByNameAndPlatform returns the bot token for the given bot name and platform.
func (h *BotHandler) GetBotTokenByNameAndPlatform(c fiber.Ctx) error {
	name := c.Get("X-Bot-Name")
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "X-Bot-Name header is required")
	}
	platform := c.Get("X-Messenger-Platform")
	if platform == "" {
		return fiber.NewError(fiber.StatusBadRequest, "X-Messenger-Platform header is required")
	}
	if platform != "telegram" && platform != "max" {
		return fiber.NewError(fiber.StatusBadRequest, "platform must be telegram or max")
	}

	token, err := h.botService.GetTokenByNameAndPlatform(c.Context(), name, platform)
	if err != nil {
		if ae, ok := apperrors.FromError(err); ok && ae.Code == fiber.StatusNotFound {
			return response.Fail(
				c,
				fiber.StatusNotFound,
				"Bot not found",
				"No bot found for name "+name+" on platform "+platform,
			)
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return response.OK(c, fiber.Map{"token": token})
}

// GetBotByName retrieves a bot by name
func (h *BotHandler) GetBotByName(c fiber.Ctx) error {
	name := c.Req().Params("name")
	if name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid bot name")
	}

	bot, err := h.botService.GetByName(c.Context(), name)
	if err != nil {
		if err.Error() == "bot not found" {
			return response.Fail(
				c,
				fiber.StatusNotFound,
				"Bot not found",
				fmt.Sprintf("Bot with name '%s' does not exist", name),
			)
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return response.OK(c, bot)
}

// UpdateBot updates an existing bot
func (h *BotHandler) UpdateBot(c fiber.Ctx) error {
	id64, err := strconv.ParseInt(c.Req().Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid bot ID")
	}

	var req dto.UpdateBotRequest

	if err := c.Bind().Body(&req); err != nil {
		return response.Fail(c, fiber.StatusBadRequest, "Invalid request body", err.Error())
	}

	if err := h.validator.Struct(req); err != nil {
		var validationErrors []dto.ValidationError
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, dto.ValidationError{
				Field:   validation.GetJSONFieldName(err.Field(), reflect.TypeOf(req)),
				Message: validatepkg.Translate(err),
			})
		}
		return response.Fail(c, fiber.StatusBadRequest, "Validation failed", validationErrors)
	}

	if req.Name == nil && req.Token == nil {
		return fiber.NewError(fiber.StatusBadRequest, "No fields to update provided")
	}

	domainReq := BotUpdateRequestToDomain(req)
	bot, err := h.botService.Update(c.Context(), int32(id64), domainReq)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return response.OK(c, BotResponseFromDomain(bot))
}

// DeleteBot removes a bot
func (h *BotHandler) DeleteBot(c fiber.Ctx) error {
	id64, err := strconv.ParseInt(c.Req().Params("id"), 10, 32)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid bot ID")
	}
	err = h.botService.Delete(c.Context(), int32(id64))
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return response.NoContent(c)
}
