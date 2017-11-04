package autorest

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

const (
	MYSQL = "mysql"
)

type QueryBuilder interface {
	CreateDSN(credentials DatabaseCredentials) string
	ParseSchema(db *sql.DB) DatabaseSchema
	BuildSelectQuery(table *Table) string
	BuildSelectAllQuery(table *Table) string
	BuildPOSTQueryAndValues(r request, t *Table) (string, []interface{})
	BuildPUTQueryAndValues(r request, t *Table) (string, []interface{})
	BuildDeleteQuery(table *Table) string
}

type DatabaseSchema map[string]*Table

type Table struct {
	Name     string
	Columns  []*Column
	PKColumn string
}

type Column struct {
	Name string
}

func (t *Table) HasColumn(colName string) bool {
	for _, col := range t.Columns {
		if col.Name == colName {
			return true
		}
	}
	return false
}

type DatabaseCredentials struct {
	Name     string
	Username string
	Password string
	Host     string
	Port     string
	Type     string
}
