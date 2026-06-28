package http

import (
	"csort.ru/classification-service/internal/dto"
	"csort.ru/classification-service/internal/httperr"
	"csort.ru/classification-service/internal/logger"
	"csort.ru/classification-service/internal/product"
	"csort.ru/classification-service/internal/transport/response"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const httpProductComponent = "transport.http.product"

type ProductHandler struct {
	service *product.ProductService
}

func NewProductHandler(service *product.ProductService) *ProductHandler {
	return &ProductHandler{
		service: service,
	}
}

func (h *ProductHandler) CreateProduct(c fiber.Ctx) error {
	log := logger.Component(c, httpProductComponent)
	var req dto.SaveProductRequest
	if err := c.Bind().Body(&req); err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid request body")
	}

	prod, err := h.service.CreateProduct(c.Context(), SaveProductRequestToDomain(req))
	if err != nil {
		return err
	}

	log.Debug().Str("product_id", prod.ID.String()).Msg("Product created successfully")
	return response.Created(c, ProductPtrToResponse(prod))
}

func (h *ProductHandler) GetProduct(c fiber.Ctx) error {
	log := logger.Component(c, httpProductComponent)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid product ID")
	}

	prod, err := h.service.GetProductByID(c.Context(), id)
	if err != nil {
		return err
	}

	log.Debug().Str("product_id", prod.ID.String()).Msg("Product retrieved successfully")
	return response.OK(c, ProductPtrToResponse(prod))
}

func (h *ProductHandler) ListProducts(c fiber.Ctx) error {
	log := logger.Component(c, httpProductComponent)
	products, err := h.service.GetAllProducts(c.Context())
	if err != nil {
		return err
	}

	log.Debug().Int("products_count", len(products)).Msg("Products retrieved successfully")
	return response.OK(c, ProductsToResponse(products))
}

func (h *ProductHandler) UpdateProduct(c fiber.Ctx) error {
	log := logger.Component(c, httpProductComponent)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid product ID")
	}

	var req dto.SaveProductRequest
	if err := c.Bind().Body(&req); err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid request body")
	}

	prod, err := h.service.UpdateProduct(c.Context(), id, SaveProductRequestToDomain(req))
	if err != nil {
		return err
	}

	log.Debug().Str("product_id", prod.ID.String()).Msg("Product updated successfully")
	return response.OK(c, ProductPtrToResponse(prod))
}

func (h *ProductHandler) DeleteProduct(c fiber.Ctx) error {
	log := logger.Component(c, httpProductComponent)
	idStr := c.Params("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return httperr.Wrap(err, fiber.StatusBadRequest, "Invalid product ID")
	}

	err = h.service.DeleteProduct(c.Context(), id)
	if err != nil {
		return err
	}

	log.Debug().Str("product_id", id.String()).Msg("Product deleted successfully")
	return response.OK(c, dto.MessageResponse{Message: "Product deleted successfully"})
}
