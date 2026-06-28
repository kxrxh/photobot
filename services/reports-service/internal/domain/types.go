package domain

type ErrorResponse struct {
	Error string `json:"error"`
}

type ReportSuccessResponse struct {
	Success    bool   `json:"success"`
	AnalysisID string `json:"analysisId"`
	Message    string `json:"message"`
}

type PresignedDownloadResponse struct {
	URL              string `json:"url"`
	ExpiresInSeconds int64  `json:"expiresInSeconds"`
}

type ReportResult struct {
	Success    bool        `json:"success"`
	AnalysisID string      `json:"analysisId"`
	UserID     int64       `json:"userId"`
	Files      ReportFiles `json:"files"`
	Error      string      `json:"error,omitempty"`
	StatusCode int         `json:"statusCode,omitempty"`
}

type ReportFiles struct {
	CSV string `json:"csv"`
	PDF string `json:"pdf"`
}

type MinIOFileStatus struct {
	Exists   bool
	Metadata map[string]string
}
