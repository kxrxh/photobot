package apierrors

import "net/http"

func BadRequest(msg string) error {
	return New(http.StatusBadRequest, msg)
}

func BadRequestf(format string, args ...any) error {
	return Newf(http.StatusBadRequest, format, args...)
}

func Unauthorized(msg string) error {
	return New(http.StatusUnauthorized, msg)
}

func Forbidden(msg string) error {
	return New(http.StatusForbidden, msg)
}

func NotFound(msg string) error {
	return New(http.StatusNotFound, msg)
}

func NotFoundf(format string, args ...any) error {
	return Newf(http.StatusNotFound, format, args...)
}

func Internal(msg string) error {
	return New(http.StatusInternalServerError, msg)
}

func InternalWrap(err error, msg string) error {
	return Wrap(err, http.StatusInternalServerError, msg)
}

func BadGatewayWrap(err error, msg string) error {
	return Wrap(err, http.StatusBadGateway, msg)
}

func UpstreamFailed(err error) error {
	if err == nil {
		return nil
	}
	return Wrap(
		err,
		http.StatusBadGateway,
		"failed to recalculate analysis with external service",
	)
}

func TooManyRequests(msg string) error {
	return New(http.StatusTooManyRequests, msg)
}
