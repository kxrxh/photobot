package bot

import "time"

type Bot struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Platform  string    `json:"platform"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Response struct {
	ID        int32     `json:"id"`
	Name      string    `json:"name"`
	Platform  string    `json:"platform"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateRequest struct {
	Name     string `json:"name"     validate:"required,min=3,max=100"`
	Token    string `json:"token"    validate:"required,min=20,max=200"`
	Platform string `json:"platform" validate:"omitempty,oneof=telegram max"`
}

type UpdateRequest struct {
	Name     *string `json:"name"     validate:"omitempty,min=3,max=100"`
	Token    *string `json:"token"    validate:"omitempty,min=20,max=200"`
	Platform *string `json:"platform" validate:"omitempty,oneof=telegram max"`
}
