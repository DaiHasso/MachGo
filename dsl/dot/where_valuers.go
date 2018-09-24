// Package dot contains mostly a set of shims for dot importing MachGo DSL.
package dot

import (
	"database/sql/driver"

	"github.com/DaiHasso/MachGo"
	"github.com/DaiHasso/MachGo/dsl"
)

// Const will create a constant value for a where clause based on the value(s)
// provided.
func Const(values ...driver.Value) dsl.WhereValuer {
	return dsl.Const(values...)
}

// ObjectColumn takes the target object and either a property name
// (ex: "FooBar") or a tag value (ex: "foo_bar") for a column on the target
// object and returns a dsl.WhereValuer to be used in a where condition.
func ObjectColumn(obj MachGo.Object, column string) dsl.WhereValuer {
	return dsl.ObjectColumn2(obj, column)
}
