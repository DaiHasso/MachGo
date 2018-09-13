package dsl

import (
	"database/sql/driver"
	"fmt"
)

// WhereValue is a value in a where clause such as a const or a column
// expression.
type WhereValue interface {
	driver.Valuer

	// Raw indicates wether the value returned is a direct expression such as
	// a.foo or a constant value such as 6.
	Raw() bool
	QueryValue(*QuerySequence) (string, []interface{})
}

// NamespacedColumn is a column that's optionally namespaced.
type NamespacedColumn struct {
	isNamespaced bool
	columnName,
	tableAlias,
	tableNamespace string
}

// QueryValue will return the value relative to the QuerySequence.
func (self NamespacedColumn) QueryValue(qs *QuerySequence) (
	string,
	[]interface{},
) {
	if self.isNamespaced {
		tableString := self.tableNamespace
		if alias, ok := qs.tableAliasMap[tableString]; ok {
			tableString = alias
		}

		return fmt.Sprintf("%s.%s", tableString, self.columnName), nil
	}

	return self.columnName, nil
}

// Value will prepare the value for the database.
func (self NamespacedColumn) Value() (driver.Value, error) {
	if self.isNamespaced {
		tableString := self.tableNamespace
		if self.tableAlias != "" {
			tableString = self.tableAlias
		}

		return fmt.Sprintf("%s.%s", tableString, self.columnName), nil
	}

	return self.columnName, nil
}

// Raw is true in this case because it is an expression not a constant.
func (self NamespacedColumn) Raw() bool {
	return true
}

// ConstantValue is a straight value.
type ConstantValue struct {
	values []driver.Value
}

// Raw is false in this case because the value is a constant.
func (self ConstantValue) Raw() bool {
	return false
}

// Value just mirrors the Value of the internal variable.
func (self ConstantValue) Value() (driver.Value, error) {
	if len(self.values) == 1 {
		return self.values[0], nil
	}
	return self.values, nil
}

// QueryValue returns a bindvar and the interface value.
func (self ConstantValue) QueryValue(qs *QuerySequence) (string, []interface{}) {
	values := make([]interface{}, len(self.values))

	var bindvars string
	for i, value := range self.values {
		if len(bindvars) > 1 {
			bindvars += ", "
		}
		bindvars += "?"
		values[i] = value
	}
	return bindvars, values
}
