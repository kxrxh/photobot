package httputil

import (
	"encoding/base64"
	"net/http"
	"strings"
)

func BytesToDataURL(body []byte, contentType string) string {
	ct := strings.TrimSpace(contentType)
	if ct == "" {
		ct = http.DetectContentType(body)
	}
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = strings.TrimSpace(ct[:i])
	}
	if ct == "" || !strings.HasPrefix(ct, "image/") {
		ct = "image/jpeg"
	}
	return "data:" + ct + ";base64," + base64.StdEncoding.EncodeToString(body)
}
