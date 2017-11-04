package autorest

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type Handler struct {
	db           *sql.DB
	tables       DatabaseSchema
	queryBuilder QueryBuilder
}

func NewHandler(credentials DatabaseCredentials) *Handler {
	handler := &Handler{}
	handler.getQueryBuilder(credentials)
	handler.connectToDB(credentials)
	handler.getDBSchema()
	return handler
}

func (handler *Handler) getQueryBuilder(credentials DatabaseCredentials) {
	switch credentials.Type {
	case MYSQL:
		handler.queryBuilder = &MysqlQueryBuilder{}
	default:
		panic("Unknown database type. Supported type is only 'mysql' right now")
	}
}

func (handler *Handler) connectToDB(credentials DatabaseCredentials) {
	dsn := handler.queryBuilder.CreateDSN(credentials)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	handler.db = db
}

func (handler *Handler) getDBSchema() {
	handler.tables = handler.queryBuilder.ParseSchema(handler.db)
}

func (h *Handler) HandleRequest(r request) (interface{}, error) {
	if !h.HasTable(r.Table) {
		return nil, ApiError{NOT_FOUND}
	}
	switch r.Action {
	case GET:
		return h.Get(r)
	case GET_ALL:
		return h.GetAll(r)
	case POST:
		return h.Post(r)
	case PUT:
		return h.Put(r)
	case DELETE:
		return "", h.Delete(r)
	default:
		return nil, ApiError{METHOD_NOT_SUPPORTED}
	}
}

func (handler *Handler) HasTable(tableName string) bool {
	_, ok := handler.tables[tableName]
	return ok
}

func (handler *Handler) GetTable(tableName string) *Table {
	if table, ok := handler.tables[tableName]; ok {
		return table
	} else {
		return nil
	}
}

func (handler *Handler) Get(r request) (interface{}, error) {
	table := handler.GetTable(r.Table)
	stmt, err := handler.db.Prepare(handler.queryBuilder.BuildSelectQuery(table))
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	rows, err := stmt.Query(r.Id)
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	defer stmt.Close()
	defer rows.Close()
	result := make(map[string]interface{})
	if rows.Next() {
		row := make([]interface{}, len(columns))
		rowPointers := make([]interface{}, len(columns))
		for i := 0; i < len(columns); i++ {
			rowPointers[i] = &row[i]
		}
		if err = rows.Scan(rowPointers...); err != nil {
			return nil, ApiError{INTERNAL_SERVER_ERROR}
		}
		for i, column := range columns {
			value, err := DetermineTypeForRawValue(rowPointers[i])
			if err != nil {
				return nil, ApiError{INTERNAL_SERVER_ERROR}
			}
			result[column] = value
		}
	} else {
		return nil, ApiError{NOT_FOUND}
	}
	return result, nil
}

func (handler *Handler) GetAll(r request) (interface{}, error) {
	table := handler.GetTable(r.Table)
	stmt, err := handler.db.Prepare(handler.queryBuilder.BuildSelectAllQuery(table))
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	rows, err := stmt.Query()
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	columns, err := rows.Columns()
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	defer stmt.Close()
	defer rows.Close()
	result := make([]map[string]interface{}, 0)
	for rows.Next() {
		item := make(map[string]interface{})
		row := make([]interface{}, len(columns))
		rowPointers := make([]interface{}, len(columns))
		for i := 0; i < len(columns); i++ {
			rowPointers[i] = &row[i]
		}
		if err = rows.Scan(rowPointers...); err != nil {
			return nil, ApiError{INTERNAL_SERVER_ERROR}
		}
		for i, column := range columns {
			value, err := DetermineTypeForRawValue(rowPointers[i])
			if err != nil {
				return nil, ApiError{INTERNAL_SERVER_ERROR}
			}
			item[column] = value
		}
		result = append(result, item)
	}
	return result, nil
}

func (handler *Handler) Post(r request) (interface{}, error) {
	table := handler.GetTable(r.Table)
	query, values := handler.queryBuilder.BuildPOSTQueryAndValues(r, table)
	stmt, err := handler.db.Prepare(query)
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	result, err := stmt.Exec(values...)
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	defer stmt.Close()
	return handler.getInsertedItem(r, result)
}

func (handler *Handler) getInsertedItem(r request, result sql.Result) (interface{}, error) {
	if newId, err := result.LastInsertId(); err == nil {
		r.Id = newId
		return handler.Get(r)
	}
	pkColumn := handler.GetTable(r.Table).PKColumn
	for key, value := range r.Data {
		if key == pkColumn {
			r.Id = value.(int64)
			return handler.Get(r)
		}
	}
	return r.Data, nil
}

func (handler *Handler) Put(r request) (interface{}, error) {
	table := handler.GetTable(r.Table)
	query, values := handler.queryBuilder.BuildPUTQueryAndValues(r, table)
	stmt, err := handler.db.Prepare(query)
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	defer stmt.Close()
	_, err = stmt.Exec(values...)
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	return handler.Get(r)
}

func (handler *Handler) Delete(r request) error {
	table := handler.GetTable(r.Table)
	stmt, err := handler.db.Prepare(handler.queryBuilder.BuildDeleteQuery(table))
	if err != nil {
		return ApiError{INTERNAL_SERVER_ERROR}
	}
	defer stmt.Close()
	_, err = stmt.Exec(r.Id)
	if err != nil {
		return ApiError{INTERNAL_SERVER_ERROR}
	}
	return nil
}
