package dot

import (
    "github.com/daihasso/machgo/dsl"
)

func queryableFromInterface(
    a, b interface{}, combiner dsl.Combiner,
) dsl.Queryable {
    lhs, rhs := dsl.InterfaceToQueryable(a), dsl.InterfaceToQueryable(b)
    return dsl.NewDefaultCondition(lhs, rhs, combiner)
}

func Equal(lhs, rhs interface{}) dsl.Queryable {
    return queryableFromInterface(lhs, rhs, dsl.EqualCombiner)
}

func NotEqual(lhs, rhs interface{}) dsl.Queryable {
    return queryableFromInterface(lhs, rhs, dsl.NotEqualCombiner)
}

func GreaterThan(lhs, rhs interface{}) dsl.Queryable {
    return queryableFromInterface(lhs, rhs, dsl.GreaterThanCombiner)
}

func LessThan(lhs, rhs interface{}) dsl.Queryable {
    return queryableFromInterface(lhs, rhs, dsl.LessThanCombiner)
}

func GreaterThanEqual(lhs, rhs interface{}) dsl.Queryable {
    return queryableFromInterface(lhs, rhs, dsl.GreaterThanEqualCombiner)
}

func LessThanEqual(lhs, rhs interface{}) dsl.Queryable {
    return queryableFromInterface(lhs, rhs, dsl.LessThanEqualCombiner)
}

func In(lhs interface{}, rhs ...interface{}) dsl.Queryable {
    lhsQueryable := dsl.InterfaceToQueryable(lhs)
    queryableRHS := dsl.InterfaceToQueryableMulti(rhs...)
    rhsQueryable := dsl.NewMultiListCondition(queryableRHS...)

    return dsl.NewDefaultCondition(lhsQueryable, rhsQueryable, dsl.InCombiner)
}

func And(ins ...interface{}) dsl.Queryable {
    insQueryable := dsl.InterfaceToQueryableMulti(ins...)

    return dsl.NewMultiAndCondition(insQueryable...)
}

func Or(ins ...interface{}) dsl.Queryable {
    insQueryable := dsl.InterfaceToQueryableMulti(ins...)

    return dsl.NewMultiOrCondition(insQueryable...)
}

func Not(ins ...interface{}) dsl.Queryable {
    insQueryable := dsl.InterfaceToQueryableMulti(ins...)
    andInsQueryable := dsl.NewMultiAndCondition(insQueryable...)

    return dsl.NotCondition{
        Value: andInsQueryable,
    }
}

// Short aliases.
func Eq(lhs, rhs interface{}) dsl.Queryable {
    return Equal(lhs, rhs)
}

func Neq(lhs, rhs interface{}) dsl.Queryable {
    return NotEqual(lhs, rhs)
}

func Gt(lhs, rhs interface{}) dsl.Queryable {
    return GreaterThan(lhs, rhs)
}

func Lt(lhs, rhs interface{}) dsl.Queryable {
    return LessThan(lhs, rhs)
}

func Gte(lhs, rhs interface{}) dsl.Queryable {
    return GreaterThanEqual(lhs, rhs)
}

func Lte(lhs, rhs interface{}) dsl.Queryable {
    return LessThanEqual(lhs, rhs)
}
