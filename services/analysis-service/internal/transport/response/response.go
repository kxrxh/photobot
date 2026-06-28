package response

import (
	"csort.ru/analysis-service/internal/apierrors"
	"github.com/bytedance/sonic"
	"github.com/gofiber/fiber/v3"
)

type SuccessResponse struct {
	Success bool `json:"success"`
	Result  any  `json:"result"`
}

func OK(c fiber.Ctx, payload any) error {
	return JSON(c, fiber.StatusOK, payload)
}

func Accepted(c fiber.Ctx, payload any) error {
	return JSON(c, fiber.StatusAccepted, payload)
}

func JSON(c fiber.Ctx, statusCode int, payload any) error {
	if statusCode == fiber.StatusNoContent {
		return NoContent(c)
	}

	body, err := sonic.Marshal(SuccessResponse{
		Success: true,
		Result:  payload,
	})
	if err != nil {
		return apierrors.Internal("Failed to encode response")
	}

	c.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSONCharsetUTF8)
	return c.Status(statusCode).Send(body)
}

func NoContent(c fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
