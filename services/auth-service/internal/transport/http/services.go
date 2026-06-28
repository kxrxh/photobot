package http

import (
	"net/http"

	"csort.ru/auth-service/internal/apperrors"
	"csort.ru/auth-service/internal/dto"
	"csort.ru/auth-service/internal/service"
	"csort.ru/auth-service/internal/transport/response"
	"github.com/gofiber/fiber/v3"
)

type ServiceHandler struct {
	serviceService *service.Service
}

func NewServiceHandler(serviceService *service.Service) *ServiceHandler {
	return &ServiceHandler{
		serviceService: serviceService,
	}
}

func (h *ServiceHandler) CreateService(c fiber.Ctx) error {
	var req dto.CreateServiceRequest
	if err := c.Bind().Body(&req); err != nil {
		return err
	}

	domainReq := CreateServiceRequestToDomain(req)
	if err := h.serviceService.Create(
		c.Context(),
		domainReq.ServiceID,
		domainReq.ServiceSecret,
	); err != nil {
		return err
	}
	return response.Created(c, dto.MessageResponse{
		Message: "service created",
	})
}

func (h *ServiceHandler) ListServices(c fiber.Ctx) error {
	services, err := h.serviceService.GetAll(c.Context())
	if err != nil {
		return err
	}

	return response.OK(c, services)
}

func (h *ServiceHandler) DeleteService(c fiber.Ctx) error {
	id := c.Req().Params("service_id", "")
	if id == "" {
		return apperrors.New(http.StatusBadRequest, "service_id param is missing")
	}

	if err := h.serviceService.Delete(c.Context(), id); err != nil {
		return err
	}
	return response.OK(c, dto.MessageResponse{
		Message: "service with id " + id + " deleted",
	})
}
