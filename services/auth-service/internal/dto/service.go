package dto

type CreateServiceRequest struct {
	ServiceID     string `json:"service_id"     validate:"required,min=1,max=100"`
	ServiceSecret string `json:"service_secret" validate:"required,min=1,max=100"`
}
