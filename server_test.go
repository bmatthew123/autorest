package autorest

import (
	"net/http"
	"io/ioutil"
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
)

func runTestServer(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	handler, mock := getHandlerForTesting(t)
	s := &Server{}
	s.handler = handler
	s.RegisterHandler("/rest/something", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hi"))
})
	go s.Run(":8080")
	return handler, mock
}

func makeTestRequest(url string, t *testing.T) string {
	client := &http.Client{}
	response, err := client.Get(url)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	return string(body)
}

func TestRegisterHandler(t *testing.T) {
	handler, mock := runTestServer(t)
	response := makeTestRequest("http://localhost:8080/rest/something", t)
	if response != "hi" {
		t.Error("Did not get expected result from endpoint registered with RegisterHandler: '" + response + "'")
	}
	checkExpectationsWereMet(t, mock)
	cleanUp(handler)
}
