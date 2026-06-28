package common

import (
	"errors"
	"fmt"
	"strings"
)

const DefaultBodyPreviewLen = 512

type Error struct {
	StatusCode  int
	Method      string
	Endpoint    string
	Message     string
	Body        string
	PeerService string
	Original    error
}

func NewHTTPError(
	peerService, method, endpoint string,
	statusCode int,
	body []byte,
	message string,
	original error,
) *Error {
	if message == "" {
		message = fmt.Sprintf("upstream HTTP error (status=%d)", statusCode)
	}
	return &Error{
		StatusCode:  statusCode,
		Method:      method,
		Endpoint:    endpoint,
		Message:     message,
		Body:        string(body),
		PeerService: peerService,
		Original:    original,
	}
}

func (e *Error) Error() string {
	if e == nil {
		return "api client error"
	}
	msg := e.Message
	if msg == "" {
		msg = "api client error"
	}
	if e.Original != nil {
		msg = fmt.Sprintf("%s (cause: %v)", msg, e.Original)
	}
	parts := []string{msg}
	if e.PeerService != "" {
		parts = append(parts, "peer="+e.PeerService)
	}
	if e.StatusCode > 0 {
		parts = append(parts, fmt.Sprintf("status=%d", e.StatusCode))
	}
	if e.Method != "" {
		parts = append(parts, "method="+e.Method)
	}
	if e.Endpoint != "" {
		parts = append(parts, "endpoint="+e.Endpoint)
	}
	if preview := e.BodyPreview(DefaultBodyPreviewLen); preview != "" {
		parts = append(parts, "body="+preview)
	}
	return strings.Join(parts, " ")
}

func (e *Error) Unwrap() error { return e.Original }

func (e *Error) HTTPStatus() int { return e.StatusCode }

func (e *Error) BodyPreview(maxLen int) string {
	if e == nil || e.Body == "" {
		return ""
	}
	if maxLen <= 0 {
		maxLen = DefaultBodyPreviewLen
	}
	s := strings.TrimSpace(e.Body)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(truncated)"
}

// Extract returns the first *Error in err's unwrap chain.
func Extract(err error) *Error {
	var apiErr *Error
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}
