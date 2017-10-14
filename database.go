package autorest

import (
	_ "github.com/go-sql-driver/mysql"
)

const (
	MYSQL = "mysql"
)

type SqlDatabase interface {
	ConnectToDB(credentials DatabaseCredentials)
	ParseSchema()
	HasTable(tableName string) bool
	GetTable(tableName string) *Table
	Get(r request) (interface{}, error)
	GetAll(r request) (interface{}, error)
	Post(r request) (interface{}, error)
	Put(r request) (interface{}, error)
	Delete(r request) error
}

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
