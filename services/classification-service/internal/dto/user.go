package dto

import "github.com/google/uuid"

type SetUserActiveClassificationRequest struct {
	ClassificationID uuid.UUID `json:"classification_id" validate:"required"`
}
