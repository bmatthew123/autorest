package autorest

import (
	"strconv"
)

const (
	BAD_REQUEST           = 400
	NOT_FOUND             = 404
	METHOD_NOT_SUPPORTED  = 405
	INTERNAL_SERVER_ERROR = 500
)

type ApiError struct {
	HTTPStatusCode int
}

func (e ApiError) Error() string {
	return strconv.Itoa(e.HTTPStatusCode)
}
