package weed

import (
	"time"

	"csort.ru/coffeebot/internal/dto"
	"csort.ru/coffeebot/internal/weed/image"
	"csort.ru/coffeebot/internal/weed/stats"
)

type ListWeedsParams struct {
	dto.PaginatedRequest
	Name         string `query:"name"`
	MainGroup    string `query:"main_group"`
	MainSubgroup string `query:"main_subgroup"`
	Subgroup     string `query:"subgroup"`
	IsQuarantine *bool  `query:"is_quarantine"`
	SortOrder    string `query:"sort_order"`

	// Size
	LMin  *float32 `query:"l_min"`
	LMax  *float32 `query:"l_max"`
	WMin  *float32 `query:"w_min"`
	WMax  *float32 `query:"w_max"`
	LWMin *float32 `query:"lw_min"`
	LWMax *float32 `query:"lw_max"`
	// HSV
	HMin *float32 `query:"h_min"`
	HMax *float32 `query:"h_max"`
	SMin *float32 `query:"s_min"`
	SMax *float32 `query:"s_max"`
	VMin *float32 `query:"v_min"`
	VMax *float32 `query:"v_max"`
	// RGB
	RMin *float32 `query:"r_min"`
	RMax *float32 `query:"r_max"`
	GMin *float32 `query:"g_min"`
	GMax *float32 `query:"g_max"`
	BMin *float32 `query:"b_min"`
	BMax *float32 `query:"b_max"`
	// Other metrics
	BrtMin     *float32 `query:"brt_min"`
	BrtMax     *float32 `query:"brt_max"`
	SqSqcrlMin *float32 `query:"sq_sqcrl_min"`
	SqSqcrlMax *float32 `query:"sq_sqcrl_max"`
}

type WeedListItem struct {
	ID              int32     `json:"id"`
	Name            string    `json:"name"`
	PrimaryImageURL string    `json:"primary_image_url,omitempty"`
	Length          float32   `json:"length"`
	Width           float32   `json:"width"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	MainGroup       string    `json:"main_group,omitempty"`
	MainSubgroup    string    `json:"main_subgroup,omitempty"`
	Subgroup        string    `json:"subgroup,omitempty"`
	IsQuarantine    bool      `json:"is_quarantine"`
	Harmfulness     string    `json:"harmfulness,omitempty"`
}

type SaveWeedParams struct {
	Statistics      *stats.WeedStatistics `json:"statistics,omitempty"`
	AnalysisIDs     []string              `json:"analysis_ids,omitempty"`
	ExcludedObjects []int64               `json:"excluded_objects,omitempty"`

	Name         string  `json:"name"          validate:"required,min=2,max=255"`
	LatinName    string  `json:"latin_name"`
	Description  string  `json:"description"`
	Length       float32 `json:"length"`
	Width        float32 `json:"width"`
	MainGroup    string  `json:"main_group"`
	MainSubgroup string  `json:"main_subgroup"`
	Subgroup     string  `json:"subgroup"`
	IsQuarantine bool    `json:"is_quarantine"`
	Harmfulness  string  `json:"harmfulness"`
}

type Weed struct {
	ID          int32   `json:"id"`
	Name        string  `json:"name"`
	LatinName   string  `json:"latin_name"`
	Description string  `json:"description"`
	Length      float32 `json:"length"`
	Width       float32 `json:"width"`

	MainGroup    string    `json:"main_group"`
	MainSubgroup string    `json:"main_subgroup"`
	Subgroup     string    `json:"subgroup"`
	IsQuarantine bool      `json:"is_quarantine"`
	Harmfulness  string    `json:"harmfulness"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type WeedDetails struct {
	ID           int32                 `json:"id"`
	Name         string                `json:"name"`
	Description  string                `json:"description"`
	Length       float32               `json:"length"`
	Width        float32               `json:"width"`
	CreatedAt    time.Time             `json:"created_at"`
	UpdatedAt    time.Time             `json:"updated_at"`
	Images       []image.WeedImageURL  `json:"images"`
	Analyses     []string              `json:"analyses"`
	Statistics   *stats.WeedStatistics `json:"statistics,omitempty"`
	MainGroup    string                `json:"main_group"`
	MainSubgroup string                `json:"main_subgroup"`
	Subgroup     string                `json:"subgroup"`
	IsQuarantine bool                  `json:"is_quarantine"`
	Harmfulness  string                `json:"harmfulness"`
}
