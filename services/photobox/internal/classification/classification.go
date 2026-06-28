package classification

type PrimaryClassification struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type SecondaryClassification struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

type ClassificationSaveRequest struct {
	Name string `json:"name" validate:"required,min=3,max=100"`
}
