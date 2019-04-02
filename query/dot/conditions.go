package dot

import (
    "github.com/daihasso/machgo/query/qtypes"
)

func interfaceToQueryableMulti(ins ...interface{}) []qtypes.Queryable {
    outs := make([]qtypes.Queryable, len(ins))

    for i, in := range(ins) {
        outs[i] = qtypes.InterfaceToQueryable(in)
    }

    return outs
}

func queryableFromInterface(
    a, b interface{}, combiner qtypes.Combiner,
) qtypes.Queryable {
    lhs, rhs := qtypes.InterfaceToQueryable(a), qtypes.InterfaceToQueryable(b)
    return qtypes.NewDefaultCondition(lhs, rhs, combiner)
}

func Equal(lhs, rhs interface{}) qtypes.Queryable {
    return queryableFromInterface(lhs, rhs, qtypes.EqualCombiner)
}

func NotEqual(lhs, rhs interface{}) qtypes.Queryable {
    return queryableFromInterface(lhs, rhs, qtypes.NotEqualCombiner)
}

func GreaterThan(lhs, rhs interface{}) qtypes.Queryable {
    return queryableFromInterface(lhs, rhs, qtypes.GreaterThanCombiner)
}

func LessThan(lhs, rhs interface{}) qtypes.Queryable {
    return queryableFromInterface(lhs, rhs, qtypes.LessThanCombiner)
}

func GreaterThanEqual(lhs, rhs interface{}) qtypes.Queryable {
    return queryableFromInterface(lhs, rhs, qtypes.GreaterThanEqualCombiner)
}

func LessThanEqual(lhs, rhs interface{}) qtypes.Queryable {
    return queryableFromInterface(lhs, rhs, qtypes.LessThanEqualCombiner)
}

func In(lhs interface{}, rhs ...interface{}) qtypes.Queryable {
    lhsQueryable := qtypes.InterfaceToQueryable(lhs)
    queryableRHS := interfaceToQueryableMulti(rhs...)
    rhsQueryable := qtypes.NewMultiListCondition(queryableRHS...)

    return qtypes.NewDefaultCondition(
        lhsQueryable, rhsQueryable, qtypes.InCombiner,
    )
}

func And(ins ...interface{}) qtypes.Queryable {
    insQueryable := interfaceToQueryableMulti(ins...)

    return qtypes.NewMultiAndCondition(insQueryable...)
}

func Or(ins ...interface{}) qtypes.Queryable {
    insQueryable := interfaceToQueryableMulti(ins...)

    return qtypes.NewMultiOrCondition(insQueryable...)
}

func Not(ins ...interface{}) qtypes.Queryable {
    insQueryable := interfaceToQueryableMulti(ins...)
    andInsQueryable := qtypes.NewMultiAndCondition(insQueryable...)

    return qtypes.NotCondition{
        Value: andInsQueryable,
    }
}

// Short aliases.
func Eq(lhs, rhs interface{}) qtypes.Queryable {
    return Equal(lhs, rhs)
}

func Neq(lhs, rhs interface{}) qtypes.Queryable {
    return NotEqual(lhs, rhs)
}

func Gt(lhs, rhs interface{}) qtypes.Queryable {
    return GreaterThan(lhs, rhs)
}

func Lt(lhs, rhs interface{}) qtypes.Queryable {
    return LessThan(lhs, rhs)
}

func Gte(lhs, rhs interface{}) qtypes.Queryable {
    return GreaterThanEqual(lhs, rhs)
}

func Lte(lhs, rhs interface{}) qtypes.Queryable {
    return LessThanEqual(lhs, rhs)
}
