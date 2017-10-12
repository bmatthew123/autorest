package autorest

import (
	"database/sql"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
)

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
			w.Write([]byte("{\"message\":\"Server returned status code " + err.Error() + "\"}"))
			return
		}
		result, err := s.handler.handleRequest(request)
		if err != nil {
			statusCode, _ := strconv.Atoi(err.Error())
			w.WriteHeader(statusCode)
			w.Write([]byte("{\"message\":\"Server returned status code " + err.Error() + "\"}"))
			return
		}
		response, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte("{\"message\":\"Server returned status code 500\"}"))
			return
		}
		w.Write(response)
	})
	panic(http.ListenAndServe(":"+port, nil))
}
