package dto

import "time"

type CreateBotRequest struct {
	Name     string `json:"name"     validate:"required,min=3,max=100"`
	Token    string `json:"token"    validate:"required,min=20,max=200"`
	Platform string `json:"platform" validate:"omitempty,oneof=telegram max"`
}

type UpdateBotRequest struct {
	Name     *string `json:"name"     validate:"omitempty,min=3,max=100"`
	Token    *string `json:"token"    validate:"omitempty,min=20,max=200"`
	Platform *string `json:"platform" validate:"omitempty,oneof=telegram max"`
}

type BotResponse struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Platform  string    `json:"platform"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
