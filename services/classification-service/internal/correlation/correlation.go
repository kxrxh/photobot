package correlation

import (
	"github.com/bytedance/sonic"
)

const (
	ParameterGroupColor    = "color"
	ParameterGroupGeometry = "geometry"
	ParameterGroupMedian   = "median"
	ParameterGroupAll      = "all"
)

type CorrelationRequest struct {
	Fractions       []ObjectGroup `json:"fractions"`
	ParameterGroups []string      `json:"parameter_groups"`
}

type ObjectGroup struct {
	Name      string `json:"name"`
	ObjectIDs []int  `json:"object_ids"`
}

type CorrelationCondition struct {
	Attribute string  `json:"attribute"`
	Operator  string  `json:"operator"`
	Value     float64 `json:"value"`
}

type CorrelationBase struct {
	Name       string                 `json:"name"`
	Conditions []CorrelationCondition `json:"conditions"`
}

type ConditionTestResult struct {
	TruePositives  int     `json:"true_positives"`
	FalsePositives int     `json:"false_positives"`
	TrueNegatives  int     `json:"true_negatives"`
	FalseNegatives int     `json:"false_negatives"`
	Precision      float64 `json:"precision"`
	Recall         float64 `json:"recall"`
	Accuracy       float64 `json:"accuracy"`
	F1Score        float64 `json:"f1_score"`
}

type CorrelationWithTest struct {
	CorrelationBase
	TestResults *ConditionTestResult `json:"test_results,omitempty"`
}

func (c *CorrelationWithTest) UnmarshalJSON(data []byte) error {
	type Alias CorrelationWithTest
	aux := &struct{ *Alias }{Alias: (*Alias)(c)}
	return sonic.Unmarshal(data, &aux)
}
