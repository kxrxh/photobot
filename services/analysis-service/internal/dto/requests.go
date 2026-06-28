package dto

import (
	"encoding/json"
	"time"
)

type RequestStatus string

const (
	RequestStatusCreated                RequestStatus = "created"
	RequestStatusProcessing             RequestStatus = "processing"
	RequestStatusWaitingForConfirmation RequestStatus = "waiting_for_confirmation"
	RequestStatusCompleted              RequestStatus = "completed"
	RequestStatusFailed                 RequestStatus = "failed"
)

type GetRequestsQueryRequest struct {
	Platform *string        `query:"platform" validate:"omitempty,oneof=telegram max"`
	Status   *RequestStatus `query:"status"`
	Limit    int32          `query:"limit"    validate:"omitempty,min=1,max=100"`
	Offset   int32          `query:"offset"   validate:"omitempty,min=0"`
}

type RequestResponse struct {
	ID           string          `json:"id"`
	UserID       string          `json:"user_id"`
	Platform     string          `json:"platform"`
	Product      string          `json:"product"`
	Status       RequestStatus   `json:"status"`
	Year         string          `json:"year,omitempty"`
	MassLiter    *float64        `json:"mass_liter,omitempty"`
	Location     string          `json:"location,omitempty"`
	Mass1000     *float64        `json:"mass_1000,omitempty"`
	Mass         *float64        `json:"mass,omitempty"`
	Images       json.RawMessage `json:"images"`
	TempID       string          `json:"temp_id,omitempty"`
	ErrorMessage string          `json:"error_message,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

type GetRequestsResponse struct {
	Requests []RequestResponse `json:"requests"`
	Total    int               `json:"total"`
}

type MessageResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
