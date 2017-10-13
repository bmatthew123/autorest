package autorest

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Server struct {
	db SqlDatabase
}

func NewServer(credentials DatabaseCredentials) *Server {
	s := &Server{}
	s.db = getDB(credentials)
	return s
}

func getDB(credentials DatabaseCredentials) SqlDatabase {
	switch credentials.Type {
	case MYSQL:
		return NewMysqlDatabase(credentials)
	default:
		panic("Database type " + credentials.Type + " not supported")
	}
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
		result, err := s.handleRequest(request)
		if err != nil {
			statusCode, _ := strconv.Atoi(err.Error())
			w.WriteHeader(statusCode)
			w.Write([]byte("{\"message\":\"Server returned status code " + err.Error() + "\"}"))
			return
		}
		respond(result, w)
	})
	panic(http.ListenAndServe(":"+port, nil))
}

func (s *Server) handleRequest(r request) (interface{}, error) {
	if !s.db.HasTable(r.Table) {
		return nil, ApiError{NOT_FOUND}
	}
	switch r.Action {
	case GET:
		return s.db.Get(r)
	case GET_ALL:
		return s.db.GetAll(r)
	case POST:
		return s.db.Post(r)
	case PUT:
		return s.db.Put(r)
	case DELETE:
		return "", s.db.Delete(r)
	default:
		return nil, ApiError{METHOD_NOT_SUPPORTED}
	}
}

func respond(result interface{}, w http.ResponseWriter) {
	response, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("{\"message\":\"Server returned status code 500\"}"))
		return
	}
	w.Write(response)
}
