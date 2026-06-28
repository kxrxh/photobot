package dto

type WeedAnalysisObject struct {
	ID              int32    `json:"id"`
	WeedID          int32    `json:"weed_id"`
	AnalysesIds     []string `json:"analyses_ids"`
	ExcludedObjects []int64  `json:"excluded_objects"`
}
