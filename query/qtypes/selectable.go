package qtypes

import (
    "fmt"
   
    "github.com/pkg/errors"

    "github.com/daihasso/machgo/base"
)

type Selectable func() (SelectExpression, error)

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

func LiteralSelectable(exp string) Selectable {
    return func() (SelectExpression, error) {
        return NewSelectExpression(exp), nil
    }
}
