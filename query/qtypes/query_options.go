package qtypes

import (
    "fmt"
)

// OptionType defines query option types.
type OptionType int

const (
    UnsetOptionType OptionType = iota
    OrderByOptionType
    LimitOptionType
    OffsetOptionType
)

// QueryOption defines a special type of queryable to be used for option on a
// query such as limit.
type QueryOption interface {
    Queryable
    OptionType() OptionType
}

// OrderByOption defines an option that orders the query by
type OrderByOption struct {
    Order Queryable
}

func (self OrderByOption) fmtString(str string) string {
    return fmt.Sprintf("ORDER BY %s", str)
}

func (OrderByOption) OptionType() OptionType {
    return OrderByOptionType
}

func (self OrderByOption) String() string {
    return self.fmtString(self.Order.String())
}

func (self OrderByOption) QueryValue(
    at *AliasedTables,
) (string, []interface{}) {
    orderStr, orderVals := self.Order.QueryValue(at)
    return self.fmtString(orderStr), orderVals
}

// AddStatements appends to an OrderByOption.
func(self *OrderByOption) AddStatements(statements ...Queryable) {
    if multiCondition, ok := self.Order.(MultiCondition); ok {
        if multiCondition.Combiner == CommaCombiner {
            multiCondition.Values = append(
                multiCondition.Values, statements...,
            )

            self.Order = multiCondition
            return
        }
    }

    allStatements := append([]Queryable{self.Order}, statements...)
    self.Order = NewMultiListCondition(allStatements...)
}

// LimitOption is sets a limit to the query.
type LimitOption struct {
    Limit int
}

func (self LimitOption) fmtString() string {
    return fmt.Sprintf("LIMIT %d", self.Limit)
}

func (LimitOption) OptionType() OptionType {
    return LimitOptionType
}

func (self LimitOption) String() string {
    return self.fmtString()
}

func (self LimitOption) QueryValue(at *AliasedTables) (string, []interface{}) {
    return self.fmtString(), nil
}

// OffsetOption sets an offset to the query.
type OffsetOption struct {
    Offset int
}

func (self OffsetOption) fmtString() string {
    return fmt.Sprintf("OFFSET %d", self.Offset)
}

func (OffsetOption) OptionType() OptionType {
    return OffsetOptionType
}

func (self OffsetOption) String() string {
    return self.fmtString()
}

func (self OffsetOption) QueryValue(
    at *AliasedTables,
) (string, []interface{}) {
    return self.fmtString(), nil
}
