package base

import (
    "github.com/daihasso/machgo/refl"
)

type CustomTabler interface {
    TableName() string
}

func BaseTable(object Base) (string, error) {
    var tableName string
    var err error
    if customTabler, ok := object.(CustomTabler); ok {
        tableName = customTabler.TableName()
    } else {
        tableName, err = refl.GuessTableName(object)
    }
    return tableName, err
}
