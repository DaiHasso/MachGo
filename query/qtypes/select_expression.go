package qtypes

import (
    "github.com/daihasso/beagle"
)

var tableColumnRegex = beagle.MustRegex(
    `^(?P<table>[^\.]+)\.(?P<column>[^\.]+)$`,
)

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
