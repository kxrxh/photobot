package correlation

import (
	"context"
	"errors"

	"csort.ru/classification-service/internal/auth"
	"csort.ru/classification-service/internal/httperr"
	"github.com/gofiber/fiber/v3"
)

type CorrelationService struct {
	client       *Client
	tokenManager *auth.TokenManager
}

func NewCorrelationService(
	baseURL string,
	tokenManager *auth.TokenManager,
) *CorrelationService {
	return &CorrelationService{
		client:       NewClient(baseURL),
		tokenManager: tokenManager,
	}
}

func (s *CorrelationService) CalculateCorrelation(
	ctx context.Context,
	req *CorrelationRequest,
) ([]CorrelationWithTest, error) {
	token := s.tokenManager.GetToken()
	result, err := s.client.CalculateCorrelation(ctx, req, token)
	if err != nil {
		if errors.Is(err, ErrUnauthorized) {
			if refreshErr := s.tokenManager.RefreshToken(ctx); refreshErr != nil {
				return nil, httperr.Wrap(
					refreshErr,
					fiber.StatusInternalServerError,
					"Failed to refresh authentication token",
				)
			}
			token = s.tokenManager.GetToken()
			return s.client.CalculateCorrelation(ctx, req, token)
		}
		return nil, httperr.Wrap(
			err,
			fiber.StatusInternalServerError,
			"Failed to calculate correlation",
		)
	}
	return result, err
}
