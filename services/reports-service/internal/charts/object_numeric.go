package charts

import (
	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/calc"
)

func numericFromObject(obj analysis.Object, field string) *float64 {
	return calc.NumericFieldValue(obj, field)
}
