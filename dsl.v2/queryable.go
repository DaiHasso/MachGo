package dsl

import (
	"math/rand"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"

	dsl1 "MachGo/dsl"
)

type Queryable interface {
	fmt.Stringer

	QueryValue(*dsl1.QuerySequence) (string, []sql.NamedArg)
}

type ConstantQueryable struct {
	Values []driver.Value
}

func (self ConstantQueryable) String() string {
	stringValue := make([]string, len(self.Values))
	for i, value := range(self.Values) {
		stringValue[i] = fmt.Sprintf("%s", value)
	}

	return strings.Join(stringValue, ",")
}

func (self ConstantQueryable) QueryValue(
	*dsl1.QuerySequence,
) (string, []sql.NamedArg) {
	args := make([]sql.NamedArg, len(self.Values))
	argStrings := make([]string, len(self.Values))
	for i, value := range(self.Values) {
		// #nosec G404
		randomNumber := rand.Int()
		argName := fmt.Sprintf("const_%d", randomNumber)
		namedArg := sql.Named(argName, value)
		argStrings[i] = fmt.Sprintf("@%s", argName)
		args[i] = namedArg
	}

	queryString := strings.Join(argStrings, ",")
	return queryString, args
}

type TableColumnQueryable struct {
	columnName,
	tableName string
}

func (self TableColumnQueryable) String() string {
	return fmt.Sprintf("%s.%s", self.tableName, self.columnName)
}

func (self TableColumnQueryable) QueryValue(
	qs *dsl1.QuerySequence,
) (string, []sql.NamedArg) {
	tableName := qs.AliasForTable(self.tableName)

	return fmt.Sprintf("%s.%s", tableName, self.columnName), nil
}

type ColumnQueryable struct {
	columnName string
}

func (self ColumnQueryable) String() string {
	return self.columnName
}

func (self ColumnQueryable) QueryValue(
	*dsl1.QuerySequence,
) (string, []sql.NamedArg) {
	return self.columnName, nil
}
