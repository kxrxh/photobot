package core

type PaginatedRequest struct {
	Limit  int32 `query:"limit"  validate:"omitempty,min=1"`
	Offset int32 `query:"offset" validate:"omitempty,min=0"`
}

type PaginatedResponse[T any] struct {
	Data   []T   `json:"data"`
	Total  int64 `json:"total"`
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}
