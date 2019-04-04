package qtypes

import (
    "database/sql"
    "database/sql/driver"
    "fmt"
    "math/rand"
    "strings"

    "github.com/daihasso/machgo/base"
    "github.com/pkg/errors"
)

// ConstantValue is simply sugar on an interface.
type ConstantValue interface{}

// Queryable is an interface that represents something that can be turned into
// a string and a special context-aware string via QueryValue.
type Queryable interface {
    fmt.Stringer

    QueryValue(*AliasedTables) (string, []interface{})
}

// LiteralQueryable uses exactly the string provided. Use with extreme caution
// as SQL injection is very possible.
type LiteralQueryable struct {
    Value string
}

func (self LiteralQueryable) String() string {
    return self.Value
}

func (self LiteralQueryable) QueryValue(
    *AliasedTables,
) (string, []interface{}) {
    return self.Value, nil
}

// ConstantQueryable is a value or series of values such as numbers or strings
// that will be used in a statement.
type ConstantQueryable struct {
    Values []interface{}
}

func (self ConstantQueryable) formatValues(stringValues []string) string {
    stringValue := strings.Join(stringValues, ", ")
    if len(stringValues) > 1 {
        // If we've got a lot of values wrap them in parens.
        stringValue = fmt.Sprintf("(%s)", stringValue)
    }

    return stringValue
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

    return self.formatValues(stringValues)
}

func (self ConstantQueryable) QueryValue(
    *AliasedTables,
) (string, []interface{}) {
    args := make([]interface{}, len(self.Values))
    argStrings := make([]string, len(self.Values))
    for i, value := range(self.Values) {
        // #nosec G404
        randomNumber := rand.Int()
        argName := fmt.Sprintf("const_%d", randomNumber)
        namedArg := sql.Named(argName, value)
        argStrings[i] = fmt.Sprintf(":%s", argName)
        args[i] = namedArg
    }

    return self.formatValues(argStrings), args
}

// TableColumnQueryable is a table/column pairing that will be used in a query.
// It will automatically be aliased as appropriate by the caller.
type TableColumnQueryable struct {
    TableName,
    ColumnName string
}

func (self TableColumnQueryable) String() string {
    return fmt.Sprintf("%s.%s", self.TableName, self.ColumnName)
}

func (self TableColumnQueryable) QueryValue(
    at *AliasedTables,
) (string, []interface{}) {
    tableName := self.TableName
    if tableAlias, ok := at.AliasForTable(self.TableName); ok {
        tableName = tableAlias
    }

    return fmt.Sprintf("%s.%s", tableName, self.ColumnName), nil
}

// ColumnQueryable is a raw column with no table attached. It may result in
// ambiguousness when multiple queried tables have the same columns. Use with
// discretion.
type ColumnQueryable struct {
    ColumnName string
}

func (self ColumnQueryable) String() string {
    return self.ColumnName
}

func (self ColumnQueryable) QueryValue(
    *AliasedTables,
) (string, []interface{}) {
    return self.ColumnName, nil
}

// AscendingQueryable takes a statement and assigns an ascending direction to
// it.
type AscendingQueryable struct {
    Statement Queryable
}

func (self AscendingQueryable) fmtString(statement string) string {
    return fmt.Sprintf("%s ASC", statement)
}

func (self AscendingQueryable) String() string {
    return self.fmtString(self.Statement.String())
}

func (self AscendingQueryable) QueryValue(
    at *AliasedTables,
) (string, []interface{}) {
    stmntString, stmntValues := self.Statement.QueryValue(at)
    return self.fmtString(stmntString), stmntValues
}

// DescendingQueryable takes a statement and assigns an descending direction to
// it.
type DescendingQueryable struct {
    Statement Queryable
}

func (self DescendingQueryable) fmtString(statement string) string {
    return fmt.Sprintf("%s DESC", statement)
}

func (self DescendingQueryable) String() string {
    return self.fmtString(self.Statement.String())
}

func (self DescendingQueryable) QueryValue(
    at *AliasedTables,
) (string, []interface{}) {
    stmntString, stmntValues := self.Statement.QueryValue(at)
    return self.fmtString(stmntString), stmntValues
}

// InterfaceToQueryable takes a generic interface and creates a
// ConstantQueryable from it.
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

// ObjectColumn is a convinience function that atomatically grabs the provided
// objects column and prefixes the provided column with it.
func ObjectColumn(obj base.Base, column string) (Queryable, error) {
    tableName, err := base.BaseTable(obj)
    if err != nil {
        return nil, errors.Wrap(err, "Couldn't get object table for queryable")
    }

    return TableColumnQueryable{
        TableName: tableName,
        ColumnName: column,
    }, nil
}
