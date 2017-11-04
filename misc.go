package autorest

import (
	"errors"
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

func DetermineTypeForRawValue(value interface{}) (interface{}, error) {
	var rawValue = *(value.(*interface{}))
	switch rawValue.(type) {
	case []byte:
		return string(rawValue.([]byte)), nil
	case int8:
		return rawValue.(int8), nil
	case int:
		return rawValue.(int), nil
	case int32:
		return rawValue.(int32), nil
	case int64:
		return rawValue.(int64), nil
	case uint8:
		return rawValue.(uint8), nil
	case uint:
		return rawValue.(uint), nil
	case uint32:
		return rawValue.(uint32), nil
	case uint64:
		return rawValue.(uint64), nil
	default:
		if rawValue != nil {
			return nil, errors.New("Unable to determine a data type for this rawValue")
		}
		return nil, nil
	}
}
