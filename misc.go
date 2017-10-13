package autorest

import (
	"strconv"
	"errors"
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

func DetermineTypeForRawValue(value interface{}) (interface{}, error) {
	var rawValue = *(value.(*interface{}))
	switch rawValue.(type) {
	case []byte:
		return string(rawValue.([]byte)), nil
	case int, int32, int8, uint, uint32, uint8, int64:
		return rawValue.(int64), nil
	default:
		if rawValue != nil {
			return nil, errors.New("Unable to determine a data type for this rawValue")
		}
		return nil, nil
	}
}
