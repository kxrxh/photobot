package reports

import (
	"strconv"
	"strings"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/calc"
	reportcontext "csort.ru/reports-service/internal/context"
	"csort.ru/reports-service/internal/maputil"
	"csort.ru/reports-service/internal/numutil"
)

func hasValidPositiveNumberStr(s string) bool {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", "."))
	f, err := strconv.ParseFloat(s, 64)
	return err == nil && f > 0
}

func ShouldCalculateMassFromObjects(rc *reportcontext.ReportContext) bool {
	return !hasValidPositiveNumberStr(rc.Weight1000Grains) ||
		!hasValidPositiveNumberStr(rc.SampleMass) ||
		!hasValidPositiveNumberStr(rc.MassLiter)
}

func AppendMassCalculatedFields(rc *reportcontext.ReportContext, objects []analysis.Object) {
	setDash := func() {
		rc.Weight1000GrainsCalc = "-"
		rc.SampleMassCalc = "-"
		rc.MassLiterCalc = "-"
	}
	if len(objects) == 0 {
		setDash()
		return
	}
	md := calc.CalculateMassDerivedFromObjects(objects)
	if md.AvgMass1000 != nil {
		rc.Weight1000GrainsCalc = numutil.FormatFloat(*md.AvgMass1000, 2) + " г"
	} else {
		rc.Weight1000GrainsCalc = "-"
	}
	if md.SumMassGrams != nil {
		rc.SampleMassCalc = numutil.FormatFloat(*md.SumMassGrams, 2) + " г"
	} else {
		rc.SampleMassCalc = "-"
	}
	w1000, wok := maputil.ParseFloatString(
		strings.TrimSpace(strings.ReplaceAll(rc.Weight1000Grains, ",", ".")),
	)
	ml, mlok := maputil.ParseFloatString(
		strings.TrimSpace(strings.ReplaceAll(rc.MassLiter, ",", ".")),
	)
	if md.AvgMass1000 != nil && wok && w1000 > 0 && mlok && ml > 0 {
		scaled := ml * (*md.AvgMass1000 / w1000)
		rc.MassLiterCalc = numutil.FormatFloat(scaled, 2) + " г"
	} else {
		rc.MassLiterCalc = "-"
	}
}
