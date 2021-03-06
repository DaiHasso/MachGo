package sess

import (
    "database/sql"
    "fmt"
    "strings"

    "github.com/daihasso/machgo/base"
)

type QueryParts struct {
    Bindvars,
    ColumnNames []string
    VariableValueMap map[string]interface{}
}

func (self *QueryParts) AddColumnName(columnNames ...string) {
    self.ColumnNames = append(self.ColumnNames, columnNames...)
}

func (self *QueryParts) AddValue(values ...sql.NamedArg) {
    for _, value := range values {
        self.VariableValueMap[value.Name] = value
        self.Bindvars = append(self.Bindvars, fmt.Sprintf(":%s", value.Name))
    }
}

func (self QueryParts) AsInsert() string {
    columns := strings.Join(self.ColumnNames, ", ")
    bindvars := strings.Join(self.Bindvars, ", ")

    return fmt.Sprintf("(%s) VALUES (%s)", columns, bindvars)
}

func (self QueryParts) AsUpdate() string {
    result := ""
    for i, column := range self.ColumnNames {
        if len(result) != 0 {
            result += ", "
        }
        result += fmt.Sprintf("%s = %s", column, self.Bindvars[i])
    }

    return result
}

type ColumnFilter func(string, *sql.NamedArg) bool

func QueryPartsFromObject(
    object base.Base, filters ...ColumnFilter,
) QueryParts {
    queryParts := &QueryParts{
        VariableValueMap: make(map[string]interface{}),
    }
    processSortedNamedValues(
        object, func(columnName string, namedArg *sql.NamedArg) {
            for _, filter := range filters {
                if filter(columnName, namedArg) {
                    return
                }
            }

            queryParts.AddColumnName(columnName)
            queryParts.AddValue(*namedArg)
        },
    )

    return *queryParts
}
