package autorest

import (
	"testing"
	"github.com/DATA-DOG/go-sqlmock"
)

var USERS_COLUMNS = []string{"id", "first_name", "last_name", "age", "email_address"}

func getHandlerForTesting(t *testing.T) (*Handler, sqlmock.Sqlmock) {
	cred := DatabaseCredentials{
		Type: "mysql",
	}
	h := &Handler{}
	h.getQueryBuilder(cred)
	h.tables = getTestingSchema()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error ocurred with sqlmock %s", err)
	}
	h.db = db
	return h, mock
}

func getTestingSchema() (schema DatabaseSchema) {
	schema = make(map[string]*Table)
	schema["users"] = &Table{
		Name: "users",
		PKColumn: "id",
		Columns: []*Column{
			&Column{Name: "id"},
			&Column{Name: "first_name"},
			&Column{Name: "last_name"},
			&Column{Name: "email_address"},
			&Column{Name: "age"},
		},
	}
	schema["products"] = &Table{
		Name: "products",
		PKColumn: "id",
		Columns: []*Column{
			&Column{Name: "id"},
			&Column{Name: "name"},
			&Column{Name: "cost"},
		},
	}
	return
}

func checkKeyAndValue(t *testing.T, key string, value interface{}, data map[string]interface{}) {
	if val, ok := data[key]; !ok {
		t.Errorf("Expected value %v for key %s but key was not present", value, key)
	} else {
		if val != value {
			t.Errorf("Expected %v for key %s but got %v", value, key, val)
		}
	}
}

func checkExpectationsWereMet(t *testing.T, sqlmock sqlmock.Sqlmock) {
	if err := sqlmock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

func cleanUp(handler *Handler) {
	handler.db.Close()
}

func TestGet(t *testing.T) {
	handler, mock := getHandlerForTesting(t)
	r := request{Table: "users", Action: GET, Id: 1}
	mock.ExpectPrepare("SELECT \\* FROM users WHERE id=\\?").
		ExpectQuery().
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(USERS_COLUMNS).AddRow(1, []byte("first"), []byte("last"), 30, []byte("guy@somewhere.com")))
	rawResult, err := handler.HandleRequest(r)
	if err != nil {
		t.Errorf("An unexpected error occurred: %s", err)
	}
	result := rawResult.(map[string]interface{})
	checkKeyAndValue(t, "id", 1, result)
	checkKeyAndValue(t, "age", 30, result)
	checkKeyAndValue(t, "first_name", "first", result)
	checkKeyAndValue(t, "last_name", "last", result)
	checkKeyAndValue(t, "email_address", "guy@somewhere.com", result)
	checkExpectationsWereMet(t, mock)
	cleanUp(handler)
}

func TestGetAll(t *testing.T) {
	handler, mock := getHandlerForTesting(t)
	r := request{Table: "users", Action: GET_ALL}
	mock.ExpectPrepare("SELECT \\* FROM users").
		ExpectQuery().
		WillReturnRows(sqlmock.NewRows(USERS_COLUMNS).
			AddRow(1, []byte("first"), []byte("last"), 30, []byte("guy@somewhere.com")).
			AddRow(2, []byte("first1"), []byte("last1"), 15, nil))
	rawResult, err := handler.HandleRequest(r)
	if err != nil {
		t.Errorf("An unexpected error occurred: %s", err)
	}
	result := rawResult.([]map[string]interface{})
	checkKeyAndValue(t, "id", 1, result[0])
	checkKeyAndValue(t, "age", 30, result[0])
	checkKeyAndValue(t, "first_name", "first", result[0])
	checkKeyAndValue(t, "last_name", "last", result[0])
	checkKeyAndValue(t, "email_address", "guy@somewhere.com", result[0])
	checkKeyAndValue(t, "id", 2, result[1])
	checkKeyAndValue(t, "age", 15, result[1])
	checkKeyAndValue(t, "first_name", "first1", result[1])
	checkKeyAndValue(t, "last_name", "last1", result[1])
	checkKeyAndValue(t, "email_address", nil, result[1])
	checkExpectationsWereMet(t, mock)
	cleanUp(handler)
}

func TestPost(t *testing.T) {
	handler, mock := getHandlerForTesting(t)
	data := make(map[string]interface{})
	data["first_name"] = "first"
	data["last_name"] = "last"
	data["age"] = 30
	r := request{Table: "users", Action: POST, Data: data}
	mock.ExpectPrepare("INSERT INTO users (.+) VALUES (.+)").
		ExpectExec().
		WithArgs(30, "first", "last").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("SELECT \\* FROM users WHERE id=\\?").
		ExpectQuery().
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(USERS_COLUMNS).AddRow(1, []byte("first"), []byte("last"), 30, nil))
	rawResult, err := handler.HandleRequest(r)
	if err != nil {
		t.Errorf("An unexpected error occurred: %s", err)
	}
	result := rawResult.(map[string]interface{})
	checkKeyAndValue(t, "id", 1, result)
	checkKeyAndValue(t, "age", 30, result)
	checkKeyAndValue(t, "first_name", "first", result)
	checkKeyAndValue(t, "last_name", "last", result)
	checkKeyAndValue(t, "email_address", nil, result)
	checkExpectationsWereMet(t, mock)
	cleanUp(handler)
}

func TestPut(t *testing.T) {
	handler, mock := getHandlerForTesting(t)
	data := make(map[string]interface{})
	data["first_name"] = "first"
	data["last_name"] = "last"
	data["age"] = 30
	r := request{Table: "users", Action: PUT, Data: data, Id: 1}
	mock.ExpectPrepare("UPDATE users SET (.+) WHERE id=\\?").
		ExpectExec().
		WithArgs(30, "first", "last", 1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectPrepare("SELECT \\* FROM users WHERE id=\\?").
		ExpectQuery().
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows(USERS_COLUMNS).AddRow(1, []byte("first"), []byte("last"), 30, nil))
	rawResult, err := handler.HandleRequest(r)
	if err != nil {
		t.Errorf("An unexpected error occurred: %s", err)
	}
	result := rawResult.(map[string]interface{})
	checkKeyAndValue(t, "id", 1, result)
	checkKeyAndValue(t, "age", 30, result)
	checkKeyAndValue(t, "first_name", "first", result)
	checkKeyAndValue(t, "last_name", "last", result)
	checkKeyAndValue(t, "email_address", nil, result)
	checkExpectationsWereMet(t, mock)
	cleanUp(handler)
}

func TestDelete(t *testing.T) {
	handler, mock := getHandlerForTesting(t)
	r := request{Table: "users", Action: DELETE, Id: 1}
	mock.ExpectPrepare("DELETE FROM users WHERE id=\\?").
		ExpectExec().
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))
	_, err := handler.HandleRequest(r)
	if err != nil {
		t.Errorf("An unexpected error occurred: %s", err)
	}
	checkExpectationsWereMet(t, mock)
	cleanUp(handler)
}
