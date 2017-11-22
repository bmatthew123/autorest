package autorest

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type Server struct {
	handler *Handler
}

func NewServer(credentials DatabaseCredentials) *Server {
	s := &Server{}
	s.handler = NewHandler(credentials)
	return s
}

func (s *Server) Run(address string) {
	http.HandleFunc("/rest/", s.handleAutorestRequest)
	panic(http.ListenAndServe(address, nil))
}

func (s *Server) RunTLS(address, certFile, keyFile string) {
	http.HandleFunc("/rest/", s.handleAutorestRequest)
	panic(http.ListenAndServeTLS(address, certFile, keyFile, nil))
}

func (s *Server) handleAutorestRequest(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		request, err := parseRequest(r)
		if err != nil {
			statusCode, _ := strconv.Atoi(err.Error())
			w.WriteHeader(statusCode)
			w.Write([]byte("{\"message\":\"Server returned status code " + err.Error() + "\"}"))
			return
		}
		result, err := s.handler.HandleRequest(request)
		if err != nil {
			statusCode, _ := strconv.Atoi(err.Error())
			w.WriteHeader(statusCode)
			w.Write([]byte("{\"message\":\"Server returned status code " + err.Error() + "\"}"))
			return
		}
		s.respond(result, w)
}

func (s *Server) respond(result interface{}, w http.ResponseWriter) {
	response, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte("{\"message\":\"Server returned status code 500\"}"))
		return
	}
	w.Write(response)
}

func (s *Server) ExcludeTables(tables ...string) {
	excludedTables := make(map[string]bool)
	for _, table := range tables {
		excludedTables[table] = true
	}
	s.handler.excludedTables = excludedTables
}

func (s *Server) ServeStaticFilesFromDirectory(directory string) {
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(directory))))
}
