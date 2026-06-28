package report

import "time"

type ReportResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    *ReportData `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type ReportData struct {
	AnalysisID  string      `json:"analysisId"`
	GeneratedAt time.Time   `json:"generatedAt"`
	Links       ReportLinks `json:"_links"`
}

type ReportLinks struct {
	CSV string `json:"csv"`
	PDF string `json:"pdf"`
}

type ReportFileResponse struct {
	Success bool            `json:"success"`
	Error   string          `json:"error,omitempty"`
	Data    *ReportFileData `json:"data,omitempty"`
}

type ReportFileData struct {
	AnalysisID string `json:"analysisId"`
	FileType   string `json:"fileType"`
	Content    []byte `json:"content"`
	Size       int64  `json:"size"`
}
