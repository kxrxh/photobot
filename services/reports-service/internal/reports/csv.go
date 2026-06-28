package reports

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"sort"

	reportcontext "csort.ru/reports-service/internal/context"
)

func BuildCSV(c *reportcontext.ReportContext) ([]byte, error) {
	if c == nil {
		return buildCSVFromMap(map[string]any{})
	}
	return buildCSVFromMap(c.ToCSVMap())
}

func buildCSVFromMap(context map[string]any) ([]byte, error) {
	var keys []string
	for key := range context {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	out := &bytes.Buffer{}
	writer := csv.NewWriter(out)
	if err := writer.Write([]string{"field", "value"}); err != nil {
		return nil, err
	}
	for _, k := range keys {
		if err := writer.Write([]string{k, fmt.Sprintf("%v", context[k])}); err != nil {
			return nil, err
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
