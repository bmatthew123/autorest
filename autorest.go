package autorest

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
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
