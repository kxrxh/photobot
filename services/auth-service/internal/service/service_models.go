package service

// CreateRequest represents the request body for creating a service
type CreateRequest struct {
	ServiceID     string `json:"service_id"     validate:"required,min=1,max=100"`
	ServiceSecret string `json:"service_secret" validate:"required,min=1,max=100"`
}

// LoginRequest represents the request body for service login
type LoginRequest struct {
	ServiceID     string `json:"service_id"     validate:"required,min=1,max=100"`
	ServiceSecret string `json:"service_secret" validate:"required,min=1,max=100"`
}
