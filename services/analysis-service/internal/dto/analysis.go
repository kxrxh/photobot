package dto

import "csort.ru/analysis-service/internal/params"

const (
	DefaultLimit     = 10
	MaxLimit         = 100
	DefaultSortBy    = "date_time"
	DefaultSortOrder = "desc"
)

type GetAnalysesQueryRequest struct {
	Limit     int32  `query:"limit"      validate:"omitempty,min=1"`
	Offset    int32  `query:"offset"     validate:"omitempty,min=0"`
	Product   string `query:"product"    validate:"omitempty,min=1,max=100,alphanumunicode"`
	ID        string `query:"id"         validate:"omitempty,min=1,max=50"`
	SortBy    string `query:"sort_by"    validate:"omitempty,oneof=date_time id product"`
	SortOrder string `query:"sort_order" validate:"omitempty,oneof=asc desc"`
}

func (q *GetAnalysesQueryRequest) SetDefaults() {
	if q.Limit == 0 {
		q.Limit = DefaultLimit
	} else if q.Limit > MaxLimit {
		q.Limit = MaxLimit
	}
	if q.SortBy == "" {
		q.SortBy = DefaultSortBy
	}
	if q.SortOrder == "" {
		q.SortOrder = DefaultSortOrder
	}
}

type AnalysisResponse struct {
	ID              string         `json:"id"`
	DateTime        string         `json:"date_time"`
	Product         *string        `json:"product,omitempty"`
	UserID          int64          `json:"user_id"`
	Source          *string        `json:"source,omitempty"`
	BotMessage      *string        `json:"bot_message,omitempty"`
	FilesSource     []string       `json:"files_source"`
	FilesOutput     []string       `json:"files_output"`
	FilesSourceURLs []string       `json:"files_source_urls,omitempty"`
	FilesOutputURLs []string       `json:"files_output_urls,omitempty"`
	ScaleMmPixel    *float64       `json:"scale_mm_pixel,omitempty"`
	AnalysisParams  *params.Params `json:"analysis_params,omitempty"`
}

type AnalysisWithObjectsResponse struct {
	AnalysisResponse
	Objects []ObjectResponse `json:"objects"`
}

type CreateAnalysisResponse struct {
	RequestID string `json:"request_id"`
}

type CreateAnalysisFormFields struct {
	Product  string `json:"product"  validate:"required,max=500"`
	Bot      string `json:"bot"      validate:"required,max=200"`
	Location string `json:"location" validate:"max=500"`
	Year     string `json:"year"     validate:"required,max=16"`
}

type MergeAnalysesRequest struct {
	UserID   int64    `json:"user_id"`
	Analyses []string `json:"analyses" validate:"required,min=2,dive,min=1"`
}

type MergeAnalysesResponse struct {
	Message string `json:"message"`
}

type PaginatedAnalysesResponse struct {
	Data   []AnalysisResponse `json:"data"`
	Total  int64              `json:"total"`
	Limit  int32              `json:"limit"`
	Offset int32              `json:"offset"`
}
