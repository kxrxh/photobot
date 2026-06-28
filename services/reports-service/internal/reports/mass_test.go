package reports

import (
	"strings"
	"testing"

	"csort.ru/reports-service/internal/api/analysis"
	reportcontext "csort.ru/reports-service/internal/context"
)

func TestShouldCalculateMassFromObjectsWhenAnyFieldInvalid(t *testing.T) {
	allValid := reportcontext.ReportContext{
		Weight1000Grains: "10",
		SampleMass:       "5",
		MassLiter:        "2",
	}
	if ShouldCalculateMassFromObjects(&allValid) {
		t.Fatal("when all three are valid positive numbers, should not force object-based mass")
	}
	missingOne := reportcontext.ReportContext{
		Weight1000Grains: "-",
		SampleMass:       "5",
		MassLiter:        "2",
	}
	if !ShouldCalculateMassFromObjects(&missingOne) {
		t.Fatal("expected calculation when weight field is not a positive number")
	}
	commaDecimal := reportcontext.ReportContext{
		Weight1000Grains: "10,5",
		SampleMass:       "1",
		MassLiter:        "1",
	}
	if ShouldCalculateMassFromObjects(&commaDecimal) {
		t.Fatal("comma as decimal should parse as valid positive number")
	}
}

func TestShouldCalculateMassFromObjectsZeroOrNegative(t *testing.T) {
	rc := reportcontext.ReportContext{
		Weight1000Grains: "0",
		SampleMass:       "1",
		MassLiter:        "1",
	}
	if !ShouldCalculateMassFromObjects(&rc) {
		t.Fatal("zero weight should trigger object-based path")
	}
}

func TestAppendMassCalculatedFieldsEmptyObjects(t *testing.T) {
	rc := reportcontext.ReportContext{}
	AppendMassCalculatedFields(&rc, nil)
	if rc.Weight1000GrainsCalc != "-" || rc.SampleMassCalc != "-" || rc.MassLiterCalc != "-" {
		t.Fatalf("expected dashes for empty objects, got calc=%q sample=%q liter=%q",
			rc.Weight1000GrainsCalc, rc.SampleMassCalc, rc.MassLiterCalc)
	}
}

func TestAppendMassCalculatedFieldsDerivesMassLiterWhenInputsValid(t *testing.T) {
	m := 1000.0
	objs := []analysis.Object{{Mass1000: &m}}
	rc := reportcontext.ReportContext{
		Weight1000Grains: "1000",
		SampleMass:       "1",
		MassLiter:        "2",
	}
	AppendMassCalculatedFields(&rc, objs)
	if !strings.Contains(rc.Weight1000GrainsCalc, "1000") {
		t.Fatalf("Weight1000GrainsCalc: %q", rc.Weight1000GrainsCalc)
	}
	if !strings.Contains(rc.SampleMassCalc, "1") {
		t.Fatalf("SampleMassCalc: %q", rc.SampleMassCalc)
	}
	// scaled = ml * (avgMass / w1000) = 2 * (1000/1000) = 2
	if !strings.HasPrefix(rc.MassLiterCalc, "2") {
		t.Fatalf("MassLiterCalc: want prefix 2, got %q", rc.MassLiterCalc)
	}
}

func TestAppendMassCalculatedFieldsMassLiterDashWhenLiterInvalid(t *testing.T) {
	m := 500.0
	objs := []analysis.Object{{Mass1000: &m}}
	rc := reportcontext.ReportContext{
		Weight1000Grains: "1000",
		SampleMass:       "1",
		MassLiter:        "-",
	}
	AppendMassCalculatedFields(&rc, objs)
	if rc.MassLiterCalc != "-" {
		t.Fatalf("MassLiterCalc: want dash when liter missing, got %q", rc.MassLiterCalc)
	}
}
