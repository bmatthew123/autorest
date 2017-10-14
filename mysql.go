package autorest

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

type QueryBuilder interface {
	CreateDSN(credentials DatabaseCredentials) string
	ParseSchema(db *sql.DB) DatabaseSchema
}

type DatabaseSchema map[string]*Table

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

func (mysql *MysqlQueryBuilder) ParseSchema(db *sql.DB) DatabaseSchema {
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
		cols, pkColumn := mysql.parseColumns(db, tableName)
		schema[tableName] = &Table{Name: tableName, Columns: cols, PKColumn: pkColumn}
	}
	return schema
}

func (mysql *MysqlQueryBuilder) parseColumns(db *sql.DB, tableName string) (cols []*Column, pkCol string) {
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

func (MysqlQueryBuilder) BuildSelectAllQuery(table *Table) string {
	return "SELECT * FROM " + table.Name
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
