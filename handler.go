package autorest

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type AutorestHandler struct {
	db SqlDatabase
}

func NewMysqlHandler(db *sql.DB) *AutorestHandler {
	return &AutorestHandler{db: NewMysqlDatabase(db)}
}

func (h *AutorestHandler) handleRequest(r request) (interface{}, error) {
	if !h.db.HasTable(r.Table) {
		return nil, ApiError{NOT_FOUND}
	}
	switch r.Action {
	case GET:
		return h.db.Get(r)
	case GET_ALL:
		return h.db.GetAll(r)
	case POST:
		return h.db.Post(r)
	case PUT:
		return h.db.Put(r)
	case DELETE:
		return "", h.db.Delete(r)
	default:
		return nil, ApiError{METHOD_NOT_SUPPORTED}
	}
}
