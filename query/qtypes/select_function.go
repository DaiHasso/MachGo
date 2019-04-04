package qtypes

import (
    "fmt"
)

// FunctionType represents the SQL function type.
type FunctionType int

const (
    UnsetFunctionType = iota
    CountFunctionType
)

// SelectFunction is a special type of Queryable that invokes a SQL function.
type SelectFunction interface {
    Queryable
    FunctionType() FunctionType
}


// SelectCount uses the SQL COUNT function with the provided Expression.
type SelectCount struct {
    Expression Queryable
}

func (SelectCount) FunctionType() FunctionType {
    return CountFunctionType
}

func (self SelectCount) fmtString(str string) string {
    return fmt.Sprintf("COUNT(%s)", str)
}

func (self SelectCount) String() string {
    str := ""
    if self.Expression != nil {
        str = self.Expression.String()
    }
    return self.fmtString(str)
}

func (self SelectCount) QueryValue(
    at *AliasedTables,
) (string, []interface{}) {
    var (
        str string
        vars []interface{}
    )
    if self.Expression != nil {
        str, vars = self.Expression.QueryValue(at)
    }
    return self.fmtString(str), vars
}
