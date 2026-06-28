package dto

type MessageResponse struct {
	Message string `json:"message"`
}

type Response[T any] struct {
	Success bool `json:"success"`
	Result  T    `json:"result"`
}
