package requests

import (
	"encoding/json"
	"time"

	"csort.ru/analysis-service/internal/identity"
	"csort.ru/analysis-service/internal/repository/requests"
)

type RequestStatus string

const (
	RequestStatusCreated                RequestStatus = "created"
	RequestStatusProcessing             RequestStatus = "processing"
	RequestStatusWaitingForConfirmation RequestStatus = "waiting_for_confirmation"
	RequestStatusCompleted              RequestStatus = "completed"
	RequestStatusFailed                 RequestStatus = "failed"
)

type Request struct {
	ID             string          `json:"id"`
	UserID         string          `json:"user_id"`
	Platform       string          `json:"platform"`
	Product        string          `json:"product"`
	Status         RequestStatus   `json:"status"`
	Year           string          `json:"year,omitempty"`
	MassLiter      *float64        `json:"mass_liter,omitempty"`
	Location       string          `json:"location,omitempty"`
	Mass1000       *float64        `json:"mass_1000,omitempty"`
	Mass           *float64        `json:"mass,omitempty"`
	Images         json.RawMessage `json:"images"`
	Classification json.RawMessage `json:"classification"`
	TempID         string          `json:"temp_id,omitempty"`
	ErrorMessage   string          `json:"error_message,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

type UserPlatformPair = identity.UserPlatformPair

type GetRequestsRequest struct {
	UserID   string         `json:"user_id"            query:"user_id"  validate:"required"`
	Platform *string        `json:"platform,omitempty" query:"platform" validate:"omitempty,oneof=telegram max"`
	Status   *RequestStatus `json:"status,omitempty"   query:"status"`
	Limit    int32          `json:"limit,omitempty"    query:"limit"    validate:"omitempty,min=1,max=100"`
	Offset   int32          `json:"offset,omitempty"   query:"offset"   validate:"omitempty,min=0"`
}

type CreateRequestRequest struct {
	ID        string          `json:"id"                   validate:"required"`
	UserID    string          `json:"user_id"              validate:"required"`
	Platform  string          `json:"platform"             validate:"required,oneof=telegram max"`
	Product   string          `json:"product"              validate:"required"`
	Status    RequestStatus   `json:"status"               validate:"required"`
	Year      string          `json:"year,omitempty"`
	MassLiter *float64        `json:"mass_liter,omitempty"`
	Location  string          `json:"location,omitempty"`
	Mass1000  *float64        `json:"mass_1000"`
	Mass      *float64        `json:"mass,omitempty"`
	Images    json.RawMessage `json:"images"`
	TempID    string          `json:"temp_id,omitempty"`
}

type CreateParams struct {
	Product   string
	UserID    string
	Platform  string
	Bot       string
	Year      string
	MassLiter *float64
	Location  string
	Mass1000  *float64
	Mass      *float64
	Images    []CreateAnalysisImageFile
}

type CreateAnalysisMessage struct {
	Product        string                    `json:"product"`
	UserID         string                    `json:"user_id"`
	Platform       string                    `json:"platform"`
	Bot            string                    `json:"bot"`
	Year           string                    `json:"year"`
	MassLiter      *float64                  `json:"mass_liter"`
	Location       string                    `json:"location"`
	Mass1000       *float64                  `json:"mass_1000"`
	Mass           *float64                  `json:"mass,omitempty"`
	Images         []CreateAnalysisImageFile `json:"images"`
	Classification json.RawMessage           `json:"classification"`
	RequestID      string                    `json:"request_id"`
}

type CreateAnalysisImageFile struct {
	ID       string `json:"id"`
	ImageURL string `json:"image_url"`
}

type AnalysisRequestData struct {
	Product   string                 `json:"product"`
	UserID    string                 `json:"user_id"`
	Platform  string                 `json:"platform"`
	TempID    string                 `json:"temp_id"`
	Status    requests.RequestStatus `json:"status"`
	CreatedAt time.Time              `json:"created_at"`
}

type NotifyErrors []map[string]string

func (e *NotifyErrors) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*e = nil
		return nil
	}
	var asStrings []string
	if err := json.Unmarshal(data, &asStrings); err == nil {
		out := make([]map[string]string, 0, len(asStrings))
		for _, s := range asStrings {
			if s == "" {
				continue
			}
			out = append(out, map[string]string{"": s})
		}
		*e = out
		return nil
	}
	var asMaps []map[string]string
	if err := json.Unmarshal(data, &asMaps); err != nil {
		return err
	}
	*e = asMaps
	return nil
}

type NotifyProcessingCompletionRequest struct {
	RequestID string       `json:"request_id" validate:"required"`
	TempID    string       `json:"temp_id"`
	Success   bool         `json:"success"`
	Message   string       `json:"message"`
	Errors    NotifyErrors `json:"errors"`
}

type ConfirmAnalysisRequest struct {
	RequestID      string   `json:"request_id"          validate:"required"`
	ExcludeObjects []string `json:"excluded_object_ids" validate:"required,min=0,max=100,dive,min=1"`
}

type UserAnalysisRequestInfo struct {
	RequestID string    `json:"request_id"`
	Product   string    `json:"product"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
