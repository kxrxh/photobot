package pdfutil

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
)

func ValidateHeader(b []byte) error {
	if len(b) < 16 {
		return fmt.Errorf("pdf output too small (%d bytes)", len(b))
	}
	if !bytes.HasPrefix(b, []byte("%PDF")) {
		return errors.New("pdf output missing %PDF header (likely incomplete render)")
	}
	return nil
}

func ChromeExecPath() string {
	for _, k := range []string{"CHROMIUM_PATH", "CHROME_PATH", "CHROME_BIN"} {
		if p := strings.TrimSpace(os.Getenv(k)); p != "" {
			return p
		}
	}
	if runtime.GOOS == "darwin" {
		return "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome"
	}
	return ""
}
