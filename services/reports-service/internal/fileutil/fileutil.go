package fileutil

import (
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// RemoveWithinDir deletes path only when it resolves inside baseDir.
func RemoveWithinDir(baseDir, path string) error {
	base := baseDir
	if base == "" {
		base = os.TempDir()
	}
	baseAbs, err := filepath.Abs(filepath.Clean(base))
	if err != nil {
		return err
	}
	pathAbs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return err
	}
	rel, err := filepath.Rel(baseAbs, pathAbs)
	if err != nil {
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return errors.New("path outside base directory")
	}
	return os.Remove(pathAbs) // #nosec G703
}

// BrowserFileURL returns a file:// URL suitable for headless Chrome on the current OS.
func BrowserFileURL(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	abs = filepath.Clean(abs)
	p := filepath.ToSlash(abs)
	if runtime.GOOS == "windows" && len(p) >= 2 && p[1] == ':' {
		p = "/" + p
	}
	u := url.URL{Scheme: "file", Path: p}
	return u.String(), nil
}
