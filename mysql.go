package autorest

import (
	"database/sql"
	"fmt"
	"strings"
	_ "github.com/go-sql-driver/mysql"
)

type MysqlQueryBuilder struct{}

func (MysqlQueryBuilder) CreateDSN(credentials DatabaseCredentials) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		credentials.Username,
		credentials.Password,
		credentials.Host,
		credentials.Port,
		credentials.Name,
	)
}

func (MysqlQueryBuilder) ParseSchema(db *sql.DB) DatabaseSchema {
	schema := make(DatabaseSchema)
	stmt, err := db.Prepare("SHOW TABLES")
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
		cols, pkColumn := MysqlQueryBuilder{}.parseColumns(db, tableName)
		schema[tableName] = &Table{Name: tableName, Columns: cols, PKColumn: pkColumn}
	}
	return schema
}

func (MysqlQueryBuilder) parseColumns(db *sql.DB, tableName string) (cols []*Column, pkCol string) {
	stmt, err := db.Prepare("SELECT column_name, data_type, column_key FROM information_schema.columns WHERE table_name='" + tableName + "'")
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

func (MysqlQueryBuilder) BuildSelectQuery(table *Table) string {
	return "SELECT * FROM " + table.Name + " WHERE " + table.PKColumn + "=?"
}

func (MysqlQueryBuilder) BuildSelectAllQuery(r request, table *Table) (query string, values []interface{}) {
	query = "SELECT * FROM " + table.Name
	values = make([]interface{}, 0)
	i := 0
	for column, value := range r.QueryParameters {
		if table.HasColumn(column) && column != "sort" {
			if i > 0 {
				query += " AND "
			} else {
				query += " WHERE "
			}
			switch value.(type) {
			case int8, int16, int32, int64, uint8, uint16, uint32, uint64:
				values = append(values, value)
				query += column + " = ? "
			case string:
				values = append(values, "%" + value.(string) + "%")
				query += column + " LIKE ? "
			case []byte:
				values = append(values, "%" + string(value.([]byte)) + "%")
				query += column + " LIKE ? "
			}
			i++
		}
	}
	query += buildSortClause(r, table)
	return
}

func buildSortClause(r request, table *Table) string {
	columnString, ok := r.QueryParameters["sort"]
	if !ok {
		return ""
	}
	var columns []string
	if strings.Contains(columnString.(string), ",") {
		columns = strings.Split(columnString.(string), ",")
	} else {
		columns = make([]string, 1)
		columns[0] = columnString.(string)
	}
	sortClause := ""
	if len(columns) > 0 {
		sortClause += " ORDER BY "
	}
	for i, column := range columns {
		if i > 0 {
			sortClause += ", "
		}
		colName := column
		sortDescending := column[0] == '-'
		if sortDescending {
			colName = column[1:]
		}
		if table.HasColumn(colName) {
			sortClause += colName
			if sortDescending {
				sortClause += " DESC"
			} else {
				sortClause += " ASC"
			}
		}
	}
	return sortClause
}

func (MysqlQueryBuilder) BuildPOSTQueryAndValues(r request, t *Table) (query string, values []interface{}) {
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

func (MysqlQueryBuilder) BuildPUTQueryAndValues(r request, t *Table) (string, []interface{}) {
	query := "UPDATE " + t.Name + " SET "
	values := make([]interface{}, 0)
	i := 0
	for key, value := range r.Data {
		if t.HasColumn(key) {
			if i > 0 {
				query += ","
			}
			query += key + "=?"
			values = append(values, value)
			i++
		}
	}
	query += " WHERE " + t.PKColumn + "=?"
	values = append(values, r.Id)
	return query, values
}

func (MysqlQueryBuilder) BuildDeleteQuery(table *Table) string {
	return "DELETE FROM " + table.Name + " WHERE " + table.PKColumn + "=?"
}
