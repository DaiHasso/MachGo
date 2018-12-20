package dsl

import (
	"math/rand"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
)

type ConstantValue interface{}

type Queryable interface {
	fmt.Stringer

	QueryValue(*QuerySequence) (string, []interface{})
}

type ConstantQueryable struct {
	Values []interface{}
}

func (self ConstantQueryable) String() string {
	stringValues := make([]string, len(self.Values))
	for i, value := range(self.Values) {
		_, isString := value.(string)
		driverValue, err := driver.DefaultParameterConverter.ConvertValue(
			value,
		)
		if err == nil {
			stringValues[i] = fmt.Sprint(driverValue)

			_, isDriverString := value.(string)
			isString = isString || isDriverString
		} else {
			// Fallback on just printing the value of whatever it is.
			stringValues[i] = fmt.Sprintf("%+v", value)
		}

		if isString {
			// Make the query look right by adding single quotes around
			// strings.
			stringValues[i] = fmt.Sprintf("'%s'", stringValues[i])
		}
	}

	stringValue := strings.Join(stringValues, ", ")
	if len(stringValues) > 1 {
		// If we've got a lot of values wrap them in parens.
		stringValue = fmt.Sprintf("(%s)", stringValue)
	}

	return stringValue
}

func (self ConstantQueryable) QueryValue(
	*QuerySequence,
) (string, []interface{}) {
	args := make([]interface{}, len(self.Values))
	argStrings := make([]string, len(self.Values))
	for i, value := range(self.Values) {
		// #nosec G404
		randomNumber := rand.Int()
		argName := fmt.Sprintf("const_%d", randomNumber)
		namedArg := sql.Named(argName, value)
		argStrings[i] = fmt.Sprintf("@%s", argName)
		args[i] = namedArg
	}

	queryString := strings.Join(argStrings, ", ")

	if len(argStrings) > 1 {
		// If we've got a lot of values wrap them in parens.
		queryString = fmt.Sprintf("(%s)", queryString)
	}
	return queryString, args
}

type TableColumnQueryable struct {
	TableName,
	ColumnName string
}

func (self TableColumnQueryable) String() string {
	return fmt.Sprintf("%s.%s", self.TableName, self.ColumnName)
}

func (self TableColumnQueryable) QueryValue(
	qs *QuerySequence,
) (string, []interface{}) {
	tableName := qs.AliasForTable(self.TableName)

	return fmt.Sprintf("%s.%s", tableName, self.ColumnName), nil
}

type ColumnQueryable struct {
	ColumnName string
}

func (self ColumnQueryable) String() string {
	return self.ColumnName
}

func (self ColumnQueryable) QueryValue(
	*QuerySequence,
) (string, []interface{}) {
	return self.ColumnName, nil
}

func InterfaceToQueryable(in interface{}) Queryable {
	var out Queryable
	if queryable, ok := in.(Queryable); ok {
		out = queryable
	} else {
		out = ConstantQueryable{
			Values: []interface{}{in},
		}
	}

	return out
}

func InterfaceToQueryableMulti(ins ...interface{}) []Queryable {
	outs := make([]Queryable, len(ins))

	for i, in := range(ins) {
		outs[i] = InterfaceToQueryable(in)
	}

	return outs
}
