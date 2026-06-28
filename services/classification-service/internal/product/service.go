package product

import (
	"context"
	"database/sql"
	"errors"

	database "csort.ru/classification-service/internal/database"
	"csort.ru/classification-service/internal/httperr"
	"csort.ru/classification-service/internal/logger"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type ProductService struct {
	q      *database.Queries
	logger zerolog.Logger
}

func NewProductService(q *database.Queries) *ProductService {
	return &ProductService{
		q:      q,
		logger: logger.GetLogger("services.product"),
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req SaveProduct) (*Product, error) {
	existingProduct, _ := s.q.GetProductByName(ctx, req.Name)
	if existingProduct.ID != uuid.Nil {
		s.logger.Warn().Str("product_name", req.Name).Msg("Product with this name already exists")
		return nil, httperr.New(fiber.StatusConflict, "Product with this name already exists")
	}

	product, err := s.q.CreateProduct(ctx, req.Name)
	if err != nil {
		s.logger.Error().Err(err).Str("product_name", req.Name).Msg("Failed to create product")
		return nil, httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to create product")
	}

	s.logger.Info().
		Str("product_id", product.ID.String()).
		Str("product_name", product.Name).
		Msg("Product created")

	return &Product{
		ID:        product.ID,
		Name:      product.Name,
		UpdatedAt: product.UpdatedAt,
		CreatedAt: product.CreatedAt,
	}, nil
}

func (s *ProductService) GetProductByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	product, err := s.q.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httperr.Wrap(err, fiber.StatusNotFound, "Product not found")
		}
		s.logger.Error().Err(err).Str("product_id", id.String()).Msg("Failed to get product")
		return nil, httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to get product")
	}

	return &Product{
		ID:        product.ID,
		Name:      product.Name,
		UpdatedAt: product.UpdatedAt,
		CreatedAt: product.CreatedAt,
	}, nil
}

func (s *ProductService) GetAllProducts(ctx context.Context) ([]Product, error) {
	products, err := s.q.GetAllProducts(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("Failed to get all products")
		return nil, httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to get all products")
	}

	s.logger.Info().Int("count", len(products)).Msg("Retrieved all products")

	responses := make([]Product, 0, len(products))
	for _, product := range products {
		responses = append(responses, Product{
			ID:        product.ID,
			Name:      product.Name,
			UpdatedAt: product.UpdatedAt,
			CreatedAt: product.CreatedAt,
		})
	}

	return responses, nil
}

func (s *ProductService) UpdateProduct(
	ctx context.Context,
	id uuid.UUID,
	req SaveProduct,
) (*Product, error) {
	if existingProduct, _ := s.q.GetProductByName(
		ctx,
		req.Name,
	); existingProduct.ID != uuid.Nil &&
		existingProduct.ID != id {
		s.logger.Warn().
			Str("product_name", req.Name).
			Str("existing_id", existingProduct.ID.String()).
			Str("updating_id", id.String()).
			Msg("Product name conflict")
		return nil, httperr.New(fiber.StatusConflict, "Product with this name already exists")
	}

	product, err := s.q.UpdateProduct(ctx, database.UpdateProductParams{
		ID:   id,
		Name: req.Name,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httperr.Wrap(err, fiber.StatusNotFound, "Product not found")
		}
		s.logger.Error().
			Err(err).
			Str("product_id", id.String()).
			Str("product_name", req.Name).
			Msg("Failed to update product")
		return nil, httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to update product")
	}

	s.logger.Info().
		Str("product_id", product.ID.String()).
		Str("product_name", product.Name).
		Msg("Product updated")

	return &Product{
		ID:        product.ID,
		Name:      product.Name,
		UpdatedAt: product.UpdatedAt,
		CreatedAt: product.CreatedAt,
	}, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id uuid.UUID) error {
	err := s.q.DeleteProduct(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return httperr.Wrap(err, fiber.StatusNotFound, "Product not found")
		}
		s.logger.Error().Err(err).Str("product_id", id.String()).Msg("Failed to delete product")
		return httperr.Wrap(err, fiber.StatusInternalServerError, "Failed to delete product")
	}

	s.logger.Info().Str("product_id", id.String()).Msg("Product deleted")

	return nil
}
