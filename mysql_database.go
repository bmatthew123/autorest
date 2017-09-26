package autorest

import (
	"fmt"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
	"strings"
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
		col := Column{Name: colName, Type: getGoTypeFromSQLType(colType)}
		cols = append(cols, &col)
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

func getGoTypeFromSQLType(sqlType string) reflect.Kind {
	switch sqlType {
	case "varchar":
		return reflect.String
	case "int":
		return reflect.Int64
	default:
		panic("No idea what to do with a column of type " + sqlType)
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
	defer stmt.Close()
	defer rows.Close()
	result := make(map[string]interface{})
	if rows.Next() {
		row := make([]interface{}, len(table.Columns))
		if err = rows.Scan(row...); err != nil {
			return nil, ApiError{INTERNAL_SERVER_ERROR}
		}
		for i, value := range row {
			switch table.Columns[i].Type {
			case reflect.String:
				result[table.Columns[i].Name] = value.(string)
			case reflect.Int64:
				result[table.Columns[i].Name] = value.(int64)
			default:
				return nil, ApiError{INTERNAL_SERVER_ERROR}
			}
		}
	}
	return result, nil
}

func buildSelectQuery(table *Table) string {
	cols := make([]string, len(table.Columns))
	for i := 0; i < len(table.Columns); i++ {
		cols[i] = table.Columns[i].Name
	}
	return "SELECT " + strings.Join(cols, ",") + " FROM " + table.Name + " WHERE " + table.PKColumn + "=?"
}

func (mysql *MysqlDatabase) GetAll(r request) (interface{}, error) {
	return nil, nil
}

func (mysql *MysqlDatabase) Post(r request) (interface{}, error) {
	return nil, nil
}

func (mysql *MysqlDatabase) Put(r request) (interface{}, error) {
	return nil, nil
}

func (mysql *MysqlDatabase) Delete(r request) (interface{}, error) {
	return nil, nil
}
