package qtypes

import (
    "github.com/daihasso/beagle"
)

var tableColumnRegex = beagle.MustRegex(
    `^(?P<table>[^\.]+)\.(?P<column>[^\.]+)$`,
)

// SelectExpression represents a column/table pairing for use in a select
// statement.
type SelectExpression struct {
    withTable bool
    columnName,
    tableName string
}

func (self SelectExpression) Table() (string, bool) {
    if !self.withTable {
        return "", false
    }

    return self.tableName, true
}

func (self SelectExpression) Column() string {
    return self.columnName
}

// NewSelectExpression takes an expression in the format `a.b` and turns it
// into a SelectExpression.
func NewSelectExpression(exp string) SelectExpression {
    if match := tableColumnRegex.Match(exp); match.Matched() {
        return SelectExpression{
            withTable: true,
            columnName: match.NamedGroup("column")[0],
            tableName: match.NamedGroup("table")[0],
        }
    }

    return SelectExpression{
        withTable: false,
        columnName: exp,
        tableName: "",
    }
}
