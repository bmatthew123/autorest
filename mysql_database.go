package autorest

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

type MysqlDatabase struct {
	db     *sql.DB
	tables map[string]*Table
}

func NewMysqlDatabase(db *sql.DB) SqlDatabase {
	mysql := &MysqlDatabase{db: db, tables: make(map[string]*Table)}
	mysql.ParseSchema()
	return mysql
}

func (mysql *MysqlDatabase) ParseSchema() {
	stmt, err := mysql.db.Prepare("SHOW TABLES")
	if err != nil {
		panic(err)
	}
	rows, err := stmt.Query()
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	defer rows.Close()
	for rows.Next() {
		var tableName string
		rows.Scan(&tableName)
		cols, pkColumn := mysql.parseColumns(tableName)
		mysql.tables[tableName] = &Table{Name: tableName, Columns: cols, PKColumn: pkColumn}
	}
}

func (mysql *MysqlDatabase) parseColumns(tableName string) (cols []*Column, pkCol string) {
	stmt, err := mysql.db.Prepare("SELECT column_name, data_type, column_key FROM information_schema.columns WHERE table_name='" + tableName + "'")
	if err != nil {
		panic(err)
	}
	rows, err := stmt.Query()
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	defer rows.Close()
	cols = make([]*Column, 0)
	for rows.Next() {
		var colName string
		var colType string
		var colKey string
		rows.Scan(&colName, &colType, &colKey)
		col := Column{Name: colName}
		cols = append(cols, &col)
		if colKey == "PRI" {
			pkCol = colName
		}
	}
	return
}

func (mysql *MysqlDatabase) getPKColumn(tableName string) (pkCol string) {
	stmt, err := mysql.db.Prepare("SELECT column_name, column_key FROM information_schema.columns WHERE table_name='" + tableName + "'")
	if err != nil {
		panic(err)
	}
	rows, err := stmt.Query()
	if err != nil {
		panic(err)
	}
	defer stmt.Close()
	defer rows.Close()
	for rows.Next() {
		var colName string
		var colKey string
		rows.Scan(&colName, &colKey)
		if colKey == "PRI" {
			pkCol = colName
		}
	}
	return
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
	stmt, err := mysql.db.Prepare(buildSelectQuery(table))
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

func buildSelectQuery(table *Table) string {
	return "SELECT * FROM " + table.Name + " WHERE " + table.PKColumn + "=?"
}

func (mysql *MysqlDatabase) GetAll(r request) (interface{}, error) {
	table := mysql.GetTable(r.Table)
	stmt, err := mysql.db.Prepare("SELECT * FROM " + table.Name)
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
	query, values := buildPOSTQueryAndValues(r, table)
	stmt, err := mysql.db.Prepare(query)
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	_, err = stmt.Exec(values...)
	if err != nil {
		return nil, ApiError{INTERNAL_SERVER_ERROR}
	}
	defer stmt.Close()
	return nil, nil
}

func buildPOSTQueryAndValues(r request, t *Table) (query string, values []interface{}) {
	query = "INSERT INTO " + t.Name + " ("
	values = make([]interface{}, 0)
	valuesClause := ""
	i := 0
	for key, value := range r.Data {
		if t.HasColumn(key) {
			if i > 0 {
				query += ","
				valuesClause += ","
			}
			query += key
			valuesClause += "?"
			values = append(values, value)
			i++
		}
	}
	query += ") VALUES (" + valuesClause + ")"
	return
}

func (mysql *MysqlDatabase) Put(r request) (interface{}, error) {
	return nil, nil
}

func (mysql *MysqlDatabase) Delete(r request) error {
	table := mysql.GetTable(r.Table)
	stmt, err := mysql.db.Prepare(buildDeleteQuery(table))
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

func buildDeleteQuery(table *Table) string {
	return "DELETE FROM " + table.Name + " WHERE " + table.PKColumn + "=?"
}
