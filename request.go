package autorest

import (
	"net/http"
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
	Table  string `json:"table"`
	Action int    `json:"action"`
	Id     string `json:"id"`
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
	return request{Id: parseIdFromRequest(r), Table: parts[1], Action: method}, nil
}

func getMethod(r *http.Request) (int, error) {
	method := strings.ToUpper(r.Method)
	switch method {
	case "GET":
		if parseIdFromRequest(r) != "" {
			return GET, nil
		}
		return GET_ALL, nil
	case "POST":
		return POST, nil
	case "PUT":
		return PUT, nil
	case "DELETE":
		return DELETE, nil
	default:
		return 0, ApiError{405}
	}
}

func parseIdFromRequest(r *http.Request) string {
	parts := strings.Split(r.URL.Path, "/")[1:]
	if len(parts) < 3 {
		return ""
	}
	return parts[2]
}
