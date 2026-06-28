package image

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"csort.ru/coffeebot/internal/database"
	"csort.ru/coffeebot/internal/minio"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

const deleteParallelism = 8

type WeedImageURL struct {
	URL       string `json:"url"`
	ID        int32  `json:"id"`
	WeedID    int32  `json:"weed_id"`
	IsPrimary bool   `json:"is_primary"`
}

type Service struct {
	queries     database.Querier
	objectStore minio.ObjectStorage
	log         zerolog.Logger
}

func NewService(
	queries database.Querier,
	objectStore minio.ObjectStorage,
	zlog zerolog.Logger,
) *Service {
	return &Service{queries: queries, objectStore: objectStore, log: zlog}
}

func (s *Service) GetWeedImages(ctx context.Context, weedID int32) ([]database.WeedImage, error) {
	return s.queries.GetWeedImages(ctx, weedID)
}

func (s *Service) BulkAddWeedImages(
	ctx context.Context,
	weedID int32,
	imageKeys []string,
	isPrimary []bool,
) error {
	if len(imageKeys) == 0 {
		return nil
	}
	flags := isPrimary
	if len(flags) == 0 {
		flags = make([]bool, len(imageKeys))
	}
	if len(flags) != len(imageKeys) {
		return fmt.Errorf("isPrimary flags length %d does not match image keys length %d",
			len(flags), len(imageKeys))
	}
	return s.queries.BulkInsertWeedImages(ctx, database.BulkInsertWeedImagesParams{
		WeedID:         weedID,
		ImageKeys:      imageKeys,
		IsPrimaryFlags: flags,
	})
}

func (s *Service) AddWeedImage(
	ctx context.Context,
	weedID int32,
	imageKey string,
) (database.WeedImage, error) {
	image, err := s.queries.AddWeedImage(
		ctx,
		database.AddWeedImageParams{WeedID: weedID, ImageKey: imageKey},
	)
	if err != nil {
		return image, err
	}

	hasPrimary, err := s.queries.WeedHasPrimaryImage(ctx, weedID)
	if err != nil {
		s.log.Error().
			Err(err).
			Int32("weed_id", weedID).
			Msg("Failed to check primary weed image after adding new image")
		return image, nil
	}

	if !hasPrimary {
		if err := s.queries.SetPrimaryImage(ctx, image.ID); err != nil {
			s.log.Error().
				Err(err).
				Int32("weed_id", weedID).
				Int32("image_id", image.ID).
				Msg("Failed to set new image as primary")
		} else {
			s.log.Debug().
				Int32("weed_id", weedID).
				Int32("image_id", image.ID).
				Msg("Set new image as primary image")
		}
	}

	return image, nil
}

func (s *Service) UploadAndAddWeedImage(
	ctx context.Context,
	weedID int32,
	fileContent io.Reader,
	fileSize int64,
	contentType string,
	originalFilename string,
) (WeedImageURL, error) {
	extension := filepath.Ext(originalFilename)
	objectKey := fmt.Sprintf("weed-images/%d/%s%s", weedID, uuid.New().String(), extension)

	if err := s.objectStore.UploadObject(
		ctx,
		objectKey,
		fileContent,
		fileSize,
		contentType,
	); err != nil {
		return WeedImageURL{}, fmt.Errorf("failed to upload image to MinIO: %w", err)
	}

	dbImage, err := s.AddWeedImage(ctx, weedID, objectKey)
	if err != nil {
		if deleteErr := s.objectStore.DeleteObject(ctx, objectKey); deleteErr != nil {
			s.log.Error().
				Err(deleteErr).
				Str("object_key", objectKey).
				Msg("Failed to cleanup uploaded file after DB error")
		}
		return WeedImageURL{}, fmt.Errorf("failed to add weed image to database: %w", err)
	}

	var image WeedImageURL
	image.ID = dbImage.ID
	image.WeedID = dbImage.WeedID

	url, err := s.objectStore.GeneratePresignedURL(ctx, dbImage.ImageKey, time.Hour*24*7)
	if err != nil {
		s.log.Error().
			Err(err).
			Str("object_key", dbImage.ImageKey).
			Msg("Failed to generate presigned URL for new image")
		image.URL = ""
	} else {
		image.URL = url
	}
	return image, nil
}

func (s *Service) SetPrimaryImage(ctx context.Context, weedID int32, imageID int32) error {
	if err := s.queries.ClearPrimaryImageForWeed(ctx, weedID); err != nil {
		return fmt.Errorf("failed to clear primary image for weed %d: %w", weedID, err)
	}
	if imageID > 0 {
		return s.queries.SetPrimaryImage(ctx, imageID)
	}
	return nil
}

func (s *Service) DeleteWeedImage(ctx context.Context, imageID int32) error {
	img, err := s.queries.GetWeedImageByID(ctx, imageID)
	if err != nil {
		return fmt.Errorf("could not get image: %w", err)
	}
	if img.IsPrimary {
		if err := s.queries.UnsetPrimaryImage(ctx, img.ID); err != nil {
			return fmt.Errorf("failed to unset primary image before deleting: %w", err)
		}
	}
	if err := s.queries.DeleteWeedImage(ctx, imageID); err != nil {
		return fmt.Errorf("failed to delete image from database: %w", err)
	}
	if deleteErr := s.objectStore.DeleteObject(ctx, img.ImageKey); deleteErr != nil {
		s.log.Error().
			Err(deleteErr).
			Str("imageKey", img.ImageKey).
			Msg("Failed to delete image from MinIO storage")
	}
	return nil
}

func (s *Service) DeleteAllWeedImages(ctx context.Context, weedID int32) error {
	images, err := s.queries.GetWeedImages(ctx, weedID)
	if err != nil {
		return fmt.Errorf("failed to get images for deletion: %w", err)
	}
	if err := s.SetPrimaryImage(ctx, weedID, 0); err != nil {
		return fmt.Errorf("failed to unset primary image before deleting all images: %w", err)
	}
	if err := s.queries.DeleteAllWeedImages(ctx, weedID); err != nil {
		return fmt.Errorf("failed to delete images from database: %w", err)
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(deleteParallelism)
	for _, img := range images {
		img := img
		g.Go(func() error {
			if deleteErr := s.objectStore.DeleteObject(gctx, img.ImageKey); deleteErr != nil {
				s.log.Error().
					Err(deleteErr).
					Str("imageKey", img.ImageKey).
					Msg("Failed to delete image from MinIO storage")
			}
			return nil
		})
	}
	return g.Wait()
}
