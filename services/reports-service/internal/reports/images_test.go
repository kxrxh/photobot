package reports

import (
	"testing"

	"csort.ru/reports-service/internal/api/analysis"
	"csort.ru/reports-service/internal/calc"

	reportcontext "csort.ru/reports-service/internal/context"
)

func TestAttachAnalysisImageURLsUsesWrappedOutputURLs(t *testing.T) {
	rc := reportcontext.ReportContext{}
	ar := analysis.AnalysisResult{
		FilesOutputURLs: []string{
			"https://cdn.example/output-0.jpg",
			" ",
			"https://cdn.example/output-1.jpg",
		},
	}

	attachAnalysisImageURLs(&rc, &ar)

	if got, want := len(rc.Img2), 2; got != want {
		t.Fatalf("expected %d image urls, got %d", want, got)
	}
	if rc.Img2[0] != "https://cdn.example/output-0.jpg" ||
		rc.Img2[1] != "https://cdn.example/output-1.jpg" {
		t.Fatalf("unexpected image urls: %#v", rc.Img2)
	}
	if got, want := rc.Img2DownloadIndices, []int{0, 1}; len(got) != len(want) ||
		got[0] != want[0] ||
		got[1] != want[1] {
		t.Fatalf("unexpected download indices: %#v", got)
	}
}

func TestAttachRepresentativeImageURLsMatchesObjectIDs(t *testing.T) {
	reps := []calc.RepresentativeGroup{
		{
			ClassName: "grain",
			Representatives: []calc.RepresentativeCard{
				{ObjectID: "1"},
				{ObjectID: "2"},
			},
		},
	}
	u1 := "https://cdn.example/object-1.jpg"
	u2 := "https://cdn.example/object-2.jpg"
	objects := []analysis.Object{
		{ID: 1, ImageURL: &u1},
		{ID: 2, ImageURL: &u2},
	}

	attachRepresentativeImageURLs(reps, objects)

	if got := reps[0].Representatives[0].ImageDataURL; got != "https://cdn.example/object-1.jpg" {
		t.Fatalf("first representative image = %q", got)
	}
	if got := reps[0].Representatives[1].ImageDataURL; got != "https://cdn.example/object-2.jpg" {
		t.Fatalf("second representative image = %q", got)
	}
}
