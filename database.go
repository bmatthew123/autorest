package autorest

import (
	_ "github.com/go-sql-driver/mysql"
	"reflect"
)

type SqlDatabase interface {
	ParseSchema()
	HasTable(tableName string) bool
	GetTable(tableName string) *Table
	Get(r request) (interface{}, error)
	GetAll(r request) (interface{}, error)
	Post(r request) (interface{}, error)
	Put(r request) (interface{}, error)
	Delete(r request) (interface{}, error)
}

type Table struct {
	Name     string
	Columns  []*Column
	PKColumn string
}

type Column struct {
	Name string
	Type reflect.Kind
}
