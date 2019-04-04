package qtypes

import (
    "fmt"
   
    "github.com/pkg/errors"

    "github.com/daihasso/machgo/base"
)

// Selectable is a thing which provides a SelectExpression (and maybe an error)
type Selectable func() (SelectExpression, error)

// BaseSelectable takes a base an provides a Selectable from it.
func BaseSelectable(obj base.Base) Selectable {
    return func() (SelectExpression, error) {
        tableName, err := base.BaseTable(obj)
        if err != nil {
            return SelectExpression{}, errors.Wrapf(
                err, "Unable to get table from object '%#+v'", obj,
            )
        }
        return NewSelectExpression(fmt.Sprintf("%s.*", tableName)), nil
    }
}

// LiteralSelectable takes the exact string and creates a select expression
// from it.
func LiteralSelectable(exp string) Selectable {
    return func() (SelectExpression, error) {
        return NewSelectExpression(exp), nil
    }
}
