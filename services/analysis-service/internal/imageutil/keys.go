package imageutil

import (
	"net/url"
	"path"
	"strings"
)

func SourceKey(analysisID, filename string) string {
	return analysisID + "/source/" + normalizeFilename(filename)
}

func OutputKey(analysisID, filename string) string {
	return analysisID + "/output/" + normalizeFilename(filename)
}

func ObjectKey(analysisID, filename string) string {
	return analysisID + "/objects/" + normalizeFilename(filename)
}

func normalizeFilename(raw string) string {
	name := strings.TrimSpace(raw)
	if name == "" {
		return ""
	}
	if parsed, err := url.Parse(name); err == nil && parsed.Scheme != "" && parsed.Host != "" {
		name = parsed.Path
	}
	name = strings.Trim(name, "/")
	base := path.Base(name)
	if base == "." {
		return ""
	}
	return base
}
