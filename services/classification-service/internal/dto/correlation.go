package dto

type ObjectGroupRef struct {
	Name      string `json:"name"`
	ObjectIDs []int  `json:"object_ids"`
}

type CorrelationRequest struct {
	Fractions       []ObjectGroupRef `json:"fractions"`
	ParameterGroups []string         `json:"parameter_groups"`
}

type CorrelationConditionRef struct {
	Attribute string  `json:"attribute"`
	Operator  string  `json:"operator"`
	Value     float64 `json:"value"`
}

type ConditionTestResultRef struct {
	TruePositives  int     `json:"true_positives"`
	FalsePositives int     `json:"false_positives"`
	TrueNegatives  int     `json:"true_negatives"`
	FalseNegatives int     `json:"false_negatives"`
	Precision      float64 `json:"precision"`
	Recall         float64 `json:"recall"`
	Accuracy       float64 `json:"accuracy"`
	F1Score        float64 `json:"f1_score"`
}

type CorrelationWithTestResponse struct {
	Name        string                    `json:"name"`
	Conditions  []CorrelationConditionRef `json:"conditions"`
	TestResults *ConditionTestResultRef   `json:"test_results,omitempty"`
}
