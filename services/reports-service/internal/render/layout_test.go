package render

import (
	"path/filepath"
	"testing"
)

func TestMainHTMLPath(t *testing.T) {
	root := "/var/templates"
	got := MainHTMLPath(root)
	want := filepath.Join(root, "html", "main.html")
	if got != want {
		t.Fatalf("MainHTMLPath(%q) = %q, want %q", root, got, want)
	}
}

func TestTemplateRootFromMainPath(t *testing.T) {
	main := filepath.Join("/app", "templates", "html", "main.html")
	got := TemplateRootFromMainPath(main)
	want := filepath.Clean(filepath.Join("/app", "templates"))
	if got != want {
		t.Fatalf("TemplateRootFromMainPath(%q) = %q, want %q", main, got, want)
	}
}
