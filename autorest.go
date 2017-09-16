package autorest

import (
	"database/sql"
	"encoding/json"
	// "fmt"
	// "os"
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
	/*user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")
	name := os.Getenv("DB_NAME")
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")*/
	dsn := credentials.Username + ":" + credentials.Password + "@tcp(" + credentials.Host + ":" + credentials.Port + ")/" + credentials.Name
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	s.db = db
}

type Server struct {
	db *sql.DB
}

func NewServer(credentials DatabaseCredentials) *Server {
	s := &Server{}
	s.connectToDB(credentials)
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
		response, err := json.Marshal(request)
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
