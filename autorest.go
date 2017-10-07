package autorest

import (
	"database/sql"
	"encoding/json"
	"errors"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
)

const (
	BAD_REQUEST           = 400
	NOT_FOUND             = 404
	METHOD_NOT_SUPPORTED  = 405
	INTERNAL_SERVER_ERROR = 500
)

type DatabaseCredentials struct {
	Name     string
	Username string
	Password string
	Host     string
	Port     string
}

func (s *Server) connectToDB(credentials DatabaseCredentials) {
	dsn := credentials.Username + ":" + credentials.Password + "@tcp(" + credentials.Host + ":" + credentials.Port + ")/" + credentials.Name
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	s.db = db
}

type Server struct {
	db      *sql.DB
	handler *AutorestHandler
}

func NewServer(credentials DatabaseCredentials) *Server {
	s := &Server{}
	s.connectToDB(credentials)
	s.handler = NewMysqlHandler(s.db)
	return s
}

func (s *Server) Run(port string) {
	http.HandleFunc("/rest/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		request, err := parseRequest(r)
		if err != nil {
			statusCode, _ := strconv.Atoi(err.Error())
			w.WriteHeader(statusCode)
			w.Write([]byte("{\"error\":" + err.Error() + "}"))
			return
		}
		result, err := s.handler.handleRequest(request)
		if err != nil {
			statusCode, _ := strconv.Atoi(err.Error())
			w.WriteHeader(statusCode)
			w.Write([]byte("{\"error\":" + err.Error() + "}"))
			return
		}
		response, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("{\"error\":500}"))
			return
		}
		w.Write(response)
	})
	panic(http.ListenAndServe(":"+port, nil))
}

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
