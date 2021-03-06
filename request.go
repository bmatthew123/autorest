package autorest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

const (
	GET = iota
	GET_ALL
	POST
	PUT
	DELETE
)

type request struct {
	Table  string
	Action int
	Id     int64
	Data   map[string]interface{}
	QueryParameters map[string]interface{}
	hasId  bool
}

func parseRequest(r *http.Request) (request, error) {
	parts := strings.Split(r.URL.Path, "/")[1:]
	if len(parts) < 2 {
		return request{}, ApiError{404}
	}
	method, err := getMethod(r)
	if err != nil {
		return request{}, err
	}
	id, err, hasId := parseIdFromRequest(r)
	if err != nil {
		return request{}, err
	}
	var data map[string]interface{}
	if method == POST || method == PUT {
		data, err = parseDataFromRequest(r)
		if err != nil {
			return request{}, err
		}
	}
	queryParameters, err := parseQueryParameters(r)
	if err != nil {
		return request{}, err
	}
	return request{
		Id: id,
		Table: parts[1],
		Action: method,
		Data: data,
		QueryParameters: queryParameters,
		hasId: hasId,
	}, nil
}

func getMethod(r *http.Request) (int, error) {
	method := strings.ToUpper(r.Method)
	switch method {
	case "GET":
		if _, _, hasId := parseIdFromRequest(r); hasId {
			return GET, nil
		} else {
			return GET_ALL, nil
		}
	case "POST":
		if _, _, hasId := parseIdFromRequest(r); hasId {
			return -1, ApiError{BAD_REQUEST}
		}
		return POST, nil
	case "PUT":
		if _, _, hasId := parseIdFromRequest(r); !hasId {
			return -1, ApiError{BAD_REQUEST}
		}
		return PUT, nil
	case "DELETE":
		if _, _, hasId := parseIdFromRequest(r); !hasId {
			return -1, ApiError{BAD_REQUEST}
		}
		return DELETE, nil
	default:
		return -1, ApiError{METHOD_NOT_SUPPORTED}
	}
}

func parseIdFromRequest(r *http.Request) (int64, error, bool) {
	parts := strings.Split(r.URL.Path, "/")[1:]
	if len(parts) < 3 {
		return 0, nil, false
	}
	id, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return 0, ApiError{BAD_REQUEST}, false
	}
	return id, nil, true
}

func parseDataFromRequest(r *http.Request) (map[string]interface{}, error) {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var data map[string]interface{}
	err := decoder.Decode(&data)
	if err != nil {
		return nil, ApiError{BAD_REQUEST}
	}
	return data, nil
}

func parseQueryParameters(r *http.Request) (map[string]interface{}, error) {
	queryParameters := make(map[string]interface{})
	for key, value := range r.URL.Query() {
		queryParameters[key] = value[0]
	}
	return queryParameters, nil
}
