// Package dot contains mostly a set of shims for dot importing machgo qtypes.
package dot

import (
    "github.com/daihasso/machgo/base"
    "github.com/daihasso/machgo/query/qtypes"
)

// Const will create a constant value for a where clause based on the value(s)
// provided.
func Const(values ...interface{}) qtypes.Queryable {
    return qtypes.ConstantQueryable{
        Values: values,
    }
}

func ObjectColumn(obj base.Base, column string) qtypes.Queryable {
    q, _ := qtypes.ObjectColumn(obj, column)
    return q
}

func SelectObject(obj base.Base) qtypes.Selectable {
    return qtypes.BaseSelectable(obj)
}
