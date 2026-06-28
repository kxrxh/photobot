package service

import (
	"context"
	"errors"

	"csort.ru/auth-service/internal/apperrors"
	"csort.ru/auth-service/internal/database"
	"csort.ru/auth-service/internal/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	db     *database.Queries
	logger zerolog.Logger
}

// NewService returns a service-registry service backed by db.
func NewService(db *database.Queries) *Service {
	return &Service{
		db:     db,
		logger: logger.GetLogger("service.client"),
	}
}

// Create creates a new service with the given ID and secret.
func (s *Service) Create(ctx context.Context, serviceID string, serviceSecret string) error {
	if serviceID == "" || serviceSecret == "" {
		s.logger.Error().Msg("Missing service name or secret")
		return apperrors.New(fiber.StatusBadRequest, "missing service name or secret")
	}

	exists, err := s.db.IsServiceExists(ctx, serviceID)
	if err != nil {
		return err
	}
	if exists {
		s.logger.Error().Msg("Service already exists")
		return apperrors.New(fiber.StatusConflict, "service already exists")
	}

	hashedSecretBytes, err := bcrypt.GenerateFromPassword(
		[]byte(serviceSecret), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	hashedSecret := string(hashedSecretBytes)
	_, err = s.db.CreateService(ctx, database.CreateServiceParams{
		ServiceID:     serviceID,
		ServiceSecret: hashedSecret,
	})
	if err != nil {
		return apperrors.Wrap(err, fiber.ErrInternalServerError.Code,
			"failed to create service")
	}

	return nil
}

// Delete deletes the service with the given ID.
func (s *Service) Delete(ctx context.Context, serviceID string) error {
	if serviceID == "" {
		s.logger.Error().Msg("Missing service ID")
		return apperrors.New(fiber.StatusBadRequest, "missing service ID")
	}

	err := s.db.DeleteService(ctx, serviceID)
	if err != nil {
		return apperrors.Wrap(err, fiber.ErrInternalServerError.Code,
			"failed to delete service")
	}

	return nil
}

func (s *Service) GetAll(ctx context.Context) ([]database.ListServicesRow, error) {
	services, err := s.db.ListServices(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err, fiber.ErrInternalServerError.Code,
			"failed to get services")
	}

	return services, nil
}

func (s *Service) Get(ctx context.Context, serviceID string) (*database.Service, error) {
	if serviceID == "" {
		s.logger.Error().Msg("Missing service ID")
		return nil, apperrors.New(fiber.StatusBadRequest, "missing service ID")
	}

	service, err := s.db.GetServiceByServiceID(ctx, serviceID)
	if err != nil {
		return nil, apperrors.Wrap(err, fiber.ErrInternalServerError.Code,
			"failed to get service")
	}

	return &service, nil
}

// ValidateCredentials checks the service secret against the stored bcrypt hash.
func (s *Service) ValidateCredentials(
	ctx context.Context,
	serviceID string,
	serviceSecret string,
) error {
	if serviceID == "" || serviceSecret == "" {
		s.logger.Error().Msg("Missing client credentials")
		return apperrors.New(fiber.StatusUnauthorized, "missing client credentials")
	}

	client, err := s.db.GetServiceByServiceID(ctx, serviceID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.logger.Error().Msg("Client not found")
			return apperrors.New(fiber.StatusNotFound, "client not found")
		}
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(client.ServiceSecret), []byte(serviceSecret))
	if err != nil {
		s.logger.Error().Err(err).Msg("Invalid client credentials")
		return apperrors.New(fiber.StatusUnauthorized, "invalid client credentials")
	}

	return nil
}
