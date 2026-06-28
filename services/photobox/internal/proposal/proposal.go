package proposal

import (
	"time"

	"csort.ru/coffeebot/internal/dto"
	"csort.ru/coffeebot/internal/weed/stats"
)

type CreateProposalParams struct {
	TargetWeedID *int32 `json:"target_weed_id,omitempty" validate:"omitempty,min=1"`

	Name         string  `json:"name,omitempty"          validate:"max=500"`
	Description  string  `json:"description,omitempty"   validate:"max=10000"`
	Harmfulness  *string `json:"harmfulness,omitempty"   validate:"omitempty,max=200"`
	MainGroup    *string `json:"main_group,omitempty"    validate:"omitempty,max=200"`
	MainSubgroup *string `json:"main_subgroup,omitempty" validate:"omitempty,max=200"`
	Subgroup     *string `json:"subgroup,omitempty"      validate:"omitempty,max=200"`

	AnalysisIDs     *[]string             `json:"analysis_ids,omitempty"     validate:"omitempty,max=500,dive,min=1,max=128"`
	ExcludedObjects *[]int64              `json:"excluded_objects,omitempty" validate:"omitempty,max=100000,dive,min=0"`
	Statistics      *stats.WeedStatistics `json:"statistics,omitempty"`
}

type UpdateProposalDraftParams struct {
	Name         *string `json:"name,omitempty"          validate:"omitempty,min=1,max=500"`
	Description  *string `json:"description,omitempty"   validate:"omitempty,max=10000"`
	Harmfulness  *string `json:"harmfulness,omitempty"   validate:"omitempty,max=200"`
	MainGroup    *string `json:"main_group,omitempty"    validate:"omitempty,max=200"`
	MainSubgroup *string `json:"main_subgroup,omitempty" validate:"omitempty,max=200"`
	Subgroup     *string `json:"subgroup,omitempty"      validate:"omitempty,max=200"`

	AnalysisIDs     *[]string             `json:"analysis_ids,omitempty"     validate:"omitempty,max=500,dive,min=1,max=128"`
	ExcludedObjects *[]int64              `json:"excluded_objects,omitempty" validate:"omitempty,max=100000,dive,min=0"`
	Statistics      *stats.WeedStatistics `json:"statistics,omitempty"`
}

type ProposalActionMessageParams struct {
	Message string `json:"message" validate:"required,min=1,max=5000"`
}

type ProposalApplyParams struct {
	Note string `json:"note,omitempty" validate:"omitempty,max=5000"`
}

type PendingWeedDraft struct {
	ID           int32   `json:"id"`
	Name         string  `json:"name"`
	LatinName    string  `json:"latin_name,omitempty"`
	Description  string  `json:"description,omitempty"`
	Length       float32 `json:"length,omitempty"`
	Width        float32 `json:"width,omitempty"`
	MainGroup    string  `json:"main_group,omitempty"`
	MainSubgroup string  `json:"main_subgroup,omitempty"`
	Subgroup     string  `json:"subgroup,omitempty"`
	IsQuarantine bool    `json:"is_quarantine"`
	Harmfulness  string  `json:"harmfulness,omitempty"`
}

type PendingWeedImageURL struct {
	ID            int32  `json:"id"`
	PendingWeedID int32  `json:"pending_weed_id"`
	URL           string `json:"url"`
	IsPrimary     bool   `json:"is_primary"`
	ImageKey      string `json:"image_key,omitempty"`
}

type Proposal struct {
	ID            int32      `json:"id"`
	Status        string     `json:"status"`
	RequestBy     int64      `json:"request_by"`
	TargetWeedID  *int32     `json:"target_weed_id,omitempty"`
	ReviewedBy    *int32     `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	ReviewNotes   *string    `json:"review_notes,omitempty"`
	SubmittedAt   *time.Time `json:"submitted_at,omitempty"`
	AppliedBy     *int32     `json:"applied_by,omitempty"`
	AppliedAt     *time.Time `json:"applied_at,omitempty"`
	AppliedWeedID *int32     `json:"applied_weed_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	Draft      PendingWeedDraft      `json:"draft"`
	Images     []PendingWeedImageURL `json:"images"`
	Analyses   []string              `json:"analyses"`
	Statistics *stats.WeedStatistics `json:"statistics,omitempty"`
}

type ListProposalsParams struct {
	dto.PaginatedRequest
	Status     *string `query:"status"`
	RequestBy  *int64  `query:"request_by"`
	ReviewedBy *int32  `query:"reviewed_by"`
	SortOrder  string  `query:"sort_order"`
}

type ProposalListItem struct {
	ID            int32      `json:"id"`
	Status        string     `json:"status"`
	RequestBy     int64      `json:"request_by"`
	ReviewedBy    *int32     `json:"reviewed_by,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	SubmittedAt   *time.Time `json:"submitted_at,omitempty"`
	AppliedWeedID *int32     `json:"applied_weed_id,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`

	PendingName string `json:"pending_name"`
}
