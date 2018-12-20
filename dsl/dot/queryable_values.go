// Package dot contains mostly a set of shims for dot importing MachGo DSL.
package dot

import (
	"MachGo"
	"MachGo/dsl"
)

// Const will create a constant value for a where clause based on the value(s)
// provided.
func Const(values ...interface{}) dsl.Queryable {
	return dsl.ConstantQueryable{
		Values: values,
	}
}

func ObjectColumn(obj MachGo.Object, column string) dsl.Queryable {
	return dsl.TableColumnQueryable{
		TableName: obj.GetTableName(),
		ColumnName: column,
	}
}
