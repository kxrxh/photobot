package user

import (
	"github.com/google/uuid"
)

type SetUserClassificationRequest struct {
	UserID           int32     `json:"user_id"           validate:"required,gte=0"`
	ClassificationID uuid.UUID `json:"classification_id" validate:"required"`
}
