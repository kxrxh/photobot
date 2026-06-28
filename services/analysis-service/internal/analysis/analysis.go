package analysis

import (
	"csort.ru/analysis-service/internal/core"
	"csort.ru/analysis-service/internal/objects"
	"csort.ru/analysis-service/internal/params"
)

const (
	DefaultLimit     = 10
	MaxLimit         = 100
	DefaultSortBy    = "date_time"
	DefaultSortOrder = "desc"
)

type Analysis struct {
	ID             string
	DateTime       string
	Product        *string
	UserID         int64
	Source         *string
	BotMessage     *string
	FilesSource    []string
	FilesOutput    []string
	ScaleMmPixel   *float64
	AnalysisParams *params.Params
}

type AnalysisWithObjects struct {
	Analysis
	Objects []objects.Object
}

type AnalysisListItem struct {
	ID           string
	DateTime     string
	Product      *string
	UserID       int64
	Source       *string
	BotMessage   *string
	FilesSource  []string
	FilesOutput  []string
	ScaleMmPixel *float64
}

type GetAnalysesPaginatedRequest struct {
	core.PaginatedRequest
	Product   string `query:"product"    validate:"omitempty,min=1,max=100,alphanumunicode"`
	ID        string `query:"id"         validate:"omitempty,min=1,max=50"`
	SortBy    string `query:"sort_by"    validate:"omitempty,oneof=date_time id product"`
	SortOrder string `query:"sort_order" validate:"omitempty,oneof=asc desc"`
}

func (p *GetAnalysesPaginatedRequest) SetDefaults() {
	if p.Limit == 0 {
		p.Limit = DefaultLimit
	} else if p.Limit > MaxLimit {
		p.Limit = MaxLimit
	}
	if p.SortBy == "" {
		p.SortBy = DefaultSortBy
	}
	if p.SortOrder == "" {
		p.SortOrder = DefaultSortOrder
	}
}

type CreateAnalysisResponse struct {
	ID string `json:"id" validate:"required"`
}
