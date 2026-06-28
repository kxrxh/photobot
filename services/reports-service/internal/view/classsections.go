package view

import (
	"html/template"
	"strings"

	"csort.ru/reports-service/internal/calc"
	"csort.ru/reports-service/internal/numutil"
)

func classMetricTables(className string, classData *calc.ClassStatistics) []MetricTable {
	if !calc.HasAnyClassMetrics(classData) {
		return nil
	}
	if calc.IsDefectClassName(className) {
		return nil
	}
	var blocks []MetricTable
	if classData.L != nil || classData.W != nil || classData.T != nil || classData.SQ != nil ||
		classData.LW != nil {
		var rows []MetricRow
		rows = appendMetricRow(rows, classData.L, "L, мм")
		rows = appendMetricRow(rows, classData.W, "W, мм")
		rows = appendMetricRow(rows, classData.T, "T (Толщина), мм")
		rows = appendMetricRow(rows, classData.SQ, "SQ, мм²")
		rows = appendMetricRow(rows, classData.LW, "L/W")
		if len(rows) > 0 {
			blocks = append(blocks, MetricTable{
				Heading: "Размеры (L, W, T, SQ, L/W)",
				Rows:    rows,
			})
		}
	}
	if classData.R != nil || classData.G != nil || classData.B != nil || classData.H != nil ||
		classData.S != nil ||
		classData.V != nil {
		var rows []MetricRow
		rows = appendMetricRow(rows, classData.R, "R")
		rows = appendMetricRow(rows, classData.G, "G")
		rows = appendMetricRow(rows, classData.B, "B")
		rows = appendMetricRow(rows, classData.H, "H")
		rows = appendMetricRow(rows, classData.S, "S")
		rows = appendMetricRow(rows, classData.V, "V")
		if len(rows) > 0 {
			blocks = append(blocks, MetricTable{
				Heading: "Цвет (R, G, B, H, S, V)",
				Rows:    rows,
			})
		}
	}
	return blocks
}

func appendMetricRow(rows []MetricRow, fs *calc.ClassFieldStats, label string) []MetricRow {
	if fs == nil {
		return rows
	}
	return append(rows, MetricRow{
		Label: label,
		Min:   numutil.FormatFloat(fs.Min, 2),
		Max:   numutil.FormatFloat(fs.Max, 2),
		Avg:   numutil.FormatFloat(fs.Avg, 2),
		Med:   numutil.FormatFloat(fs.Med, 2),
	})
}

func repGroup(title string, group calc.RepresentativeGroup) (RepGroup, bool) {
	if len(group.Representatives) == 0 {
		return RepGroup{}, false
	}
	cards := make([]RepCard, 0, len(group.Representatives))
	for _, card := range group.Representatives {
		cards = append(cards, RepCard{
			HasImage:     strings.TrimSpace(card.ImageDataURL) != "",
			ImageDataURL: template.URL(card.ImageDataURL),
			ObjectID:     card.ObjectID,
			Value:        numutil.FormatFloat(card.Value, 3),
		})
	}
	return RepGroup{
		Title:        title,
		TotalObjects: group.TotalObjects,
		Cards:        cards,
	}, true
}
