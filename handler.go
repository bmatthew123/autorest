package autorest

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

const (
	BAD_REQUEST           = 400
	NOT_FOUND             = 404
	METHOD_NOT_SUPPORTED  = 405
	INTERNAL_SERVER_ERROR = 500
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
		return nil, h.db.Delete(r)
	default:
		return nil, ApiError{METHOD_NOT_SUPPORTED}
	}
}
