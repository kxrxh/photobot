package observability

import (
	"fmt"
	"net/http"
)

func HTTPStatusError(statusCode int) error {
	text := http.StatusText(statusCode)
	if text == "" {
		text = "non-OK status"
	}
	return fmt.Errorf("HTTP %d %s", statusCode, text)
}
