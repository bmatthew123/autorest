package autorest

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type MysqlDatabase struct {
	db     *sql.DB
	tables DatabaseSchema
}

func NewMysqlDatabase(credentials DatabaseCredentials) SqlDatabase {
	mysql := &MysqlDatabase{tables: make(map[string]*Table)}
	mysql.ConnectToDB(credentials)
	mysql.tables = (&MysqlQueryBuilder{}).ParseSchema(mysql.db)
	return mysql
}

func (mysql *MysqlDatabase) ConnectToDB(credentials DatabaseCredentials) {
	dsn := MysqlQueryBuilder{}.CreateDSN(credentials)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}
	mysql.db = db
}

func (mysql *MysqlDatabase) ParseSchema() {
}

func (mysql *MysqlDatabase) HasTable(tableName string) bool {
	_, ok := mysql.tables[tableName]
	return ok
}

func (mysql *MysqlDatabase) GetTable(tableName string) *Table {
	if table, ok := mysql.tables[tableName]; ok {
		return table
	} else {
		return nil
	}
}

func (mysql *MysqlDatabase) Get(r request) (interface{}, error) {
	table := mysql.GetTable(r.Table)
	stmt, err := mysql.db.Prepare(MysqlQueryBuilder{}.BuildSelectQuery(table))
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

func (mysql *MysqlDatabase) GetAll(r request) (interface{}, error) {
	table := mysql.GetTable(r.Table)
	stmt, err := mysql.db.Prepare(MysqlQueryBuilder{}.BuildSelectAllQuery(table))
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

func (mysql *MysqlDatabase) Post(r request) (interface{}, error) {
	table := mysql.GetTable(r.Table)
	query, values := MysqlQueryBuilder{}.BuildPOSTQueryAndValues(r, table)
	stmt, err := mysql.db.Prepare(query)
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	result, err := stmt.Exec(values...)
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	defer stmt.Close()
	return mysql.getInsertedItem(r, result)
}

func (mysql *MysqlDatabase) getInsertedItem(r request, result sql.Result) (interface{}, error) {
	if newId, err := result.LastInsertId(); err == nil {
		r.Id = newId
		return mysql.Get(r)
	}
	pkColumn := mysql.GetTable(r.Table).PKColumn
	for key, value := range r.Data {
		if key == pkColumn {
			r.Id = value.(int64)
			return mysql.Get(r)
		}
	}
	return r.Data, nil
}

func (mysql *MysqlDatabase) Put(r request) (interface{}, error) {
	table := mysql.GetTable(r.Table)
	query, values := MysqlQueryBuilder{}.BuildPUTQueryAndValues(r, table)
	stmt, err := mysql.db.Prepare(query)
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	defer stmt.Close()
	_, err = stmt.Exec(values...)
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	return mysql.Get(r)
}

func (mysql *MysqlDatabase) Delete(r request) error {
	table := mysql.GetTable(r.Table)
	stmt, err := mysql.db.Prepare(MysqlQueryBuilder{}.BuildDeleteQuery(table))
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
