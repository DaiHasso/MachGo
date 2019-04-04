package qtypes

import (
    "fmt"
)

type queryableValuer func(Queryable) (string, []interface{})

var aliasedTablesValuer = func(at *AliasedTables) queryableValuer {
    return func(q Queryable) (string, []interface{}) {
        return q.QueryValue(at)
    }
}
var stringValuer = func(q Queryable) (string, []interface{}) {
    return q.String(), nil
}

// ConditionEvaluator is a function that takes two queryables and runs a
// transformation function (queryableValuer) on them to return their string
// representation and any values associated.
type ConditionEvaluator func(
    queryableValuer, Queryable, Queryable,
) (string, []interface{})


// NewDefaultEvaluator returns the standard evaluator.
func NewDefaultEvaluator(
    combiner Combiner, valueModifiers ...ValueModifier,
) ConditionEvaluator {
    return func(
        v queryableValuer, lhValues, rhValues Queryable,
    ) (string, []interface{}) {
        leftQueryString, leftArgs := v(lhValues)
        rightQueryString, rightArgs := v(rhValues)
        allArgs := append(leftArgs, rightArgs...)

        for _, valueModifier := range(valueModifiers) {
            leftQueryString, rightQueryString = valueModifier(
                leftQueryString, rightQueryString,
            )
        }

        queryString := fmt.Sprintf(
            "(%s %s %s)", leftQueryString, combiner.String(), rightQueryString,
        )
        return queryString, allArgs
    }
}

// DefaultCondition represents a standard condition that just combines on
// lefthand set of values with a righthand set of values.
type DefaultCondition struct {
    LHValues,
    RHValues Queryable

    Evaluator ConditionEvaluator
}

func (self DefaultCondition) QueryValue(
    at *AliasedTables,
) (string, []interface{}) {
    return self.Evaluator(
        aliasedTablesValuer(at), self.LHValues, self.RHValues,
    )
}

func (self DefaultCondition) String() string {
    queryString, _ := self.Evaluator(
        stringValuer, self.LHValues, self.RHValues,
    )

    return queryString
}

// NewDefaultCondition creates a new default condition combining the values
// provided via the combiner provided.
func NewDefaultCondition(
    LHValues, RHValues Queryable, combiner Combiner,
) Queryable {
    var modifiers []ValueModifier
    switch(combiner) {
        case InCombiner:
        modifiers = append(modifiers, rightParenValueModifier)
    }

    return DefaultCondition{
        LHValues: LHValues,
        RHValues: RHValues,
        Evaluator: NewDefaultEvaluator(combiner, modifiers...),
    }
}

// MultiCondition combines multiple values in a serial fashion using the
// provided Combiner.
type MultiCondition struct {
    Values []Queryable
    Combiner Combiner
}

func (self MultiCondition) QueryValue(
    at *AliasedTables,
) (string, []interface{}) {
    var allArgs []interface{}
    queries := make([]string, len(self.Values))
    for i, value := range self.Values {
        queryString, args := value.QueryValue(at)
        queries[i] = queryString
        allArgs = append(allArgs, args...)
    }

    finalQueryString := self.Combiner.Join(queries...)

    return finalQueryString, allArgs
}

func (self MultiCondition) String() string {
    queries := make([]string, len(self.Values))
    for i, value := range self.Values {
        queryString := value.String()
        queries[i] = queryString
    }

    finalQueryString := self.Combiner.Join(queries...)

    return finalQueryString
}

// NotCondition creates a sql NOT on the provided statment.
type NotCondition struct {
    Value Queryable
}

func (self NotCondition) QueryValue(
    at *AliasedTables,
) (string, []interface{}) {
    valueQueryString, args := self.Value.QueryValue(at)
    queryString := fmt.Sprintf(
        "(%s %s)", NotCombiner.String(), maybeParen(valueQueryString),
    )
    return queryString, args
}

func (self NotCondition) String() string {
    valueQueryString := self.Value.String()
    queryString := fmt.Sprintf(
        "(%s %s)", NotCombiner.String(), maybeParen(valueQueryString),
    )
    return queryString
}

// NewMultiOrCondition takes the provided values and combines them in the
// fashion of `a=5 OR c=1`
func NewMultiOrCondition(values ...Queryable) Queryable {
    return MultiCondition{
        Values: values,
        Combiner: OrCombiner,
    }
}

// NewMultiAndCondition takes the provided values and combines them in the
// fashion of `a=5 AND c=1`
func NewMultiAndCondition(values ...Queryable) Queryable {
    return MultiCondition{
        Values: values,
        Combiner: AndCombiner,
    }
}

// NewMultiListCondition takes the provided values and combines them in the
// fashion of `a=5, c=1`
func NewMultiListCondition(values ...Queryable) Queryable {
    return MultiCondition{
        Values: values,
        Combiner: CommaCombiner,
    }
}
