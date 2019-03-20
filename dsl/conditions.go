package dsl

import (
    "fmt"
)

type ConditionEvaluator func(
    *QuerySequence, Queryable, Queryable,
) (string, []interface{})


func NewDefaultEvaluator(
    combiner Combiner, valueModifiers ...ValueModifier,
) ConditionEvaluator {
    return func(
        qs *QuerySequence,
        lhValues,
        rhValues Queryable,
    ) (string, []interface{}) {
        var allArgs []interface{}
        var leftQueryString, rightQueryString string

        if qs != nil {
            var leftArgs, rightArgs []interface{}
            leftQueryString, leftArgs = lhValues.QueryValue(qs)
            rightQueryString, rightArgs = rhValues.QueryValue(qs)
            allArgs = append(leftArgs, rightArgs...)
        } else {
            leftQueryString = lhValues.String()
            rightQueryString = rhValues.String()
        }

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

type DefaultCondition struct {
    LHValues,
    RHValues Queryable

    Evaluator ConditionEvaluator
}

func (self DefaultCondition) QueryValue(
    qs *QuerySequence,
) (string, []interface{}) {
    return self.Evaluator(qs, self.LHValues, self.RHValues)
}

func (self DefaultCondition) String() string {
    queryString, _ := self.Evaluator(nil, self.LHValues, self.RHValues)

    return queryString
}

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

type MultiCondition struct {
    Values []Queryable
    Combiner Combiner
}

func (self MultiCondition) QueryValue(
    qs *QuerySequence,
) (string, []interface{}) {
    var allArgs []interface{}
    queries := make([]string, len(self.Values))
    for i, value := range self.Values {
        queryString, args := value.QueryValue(qs)
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

type NotCondition struct {
    Value Queryable
}

func (self NotCondition) QueryValue(
    qs *QuerySequence,
) (string, []interface{}) {
    valueQueryString, args := self.Value.QueryValue(qs)
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

func NewMultiOrCondition(values ...Queryable) Queryable {
    return MultiCondition{
        Values: values,
        Combiner: OrCombiner,
    }
}

func NewMultiAndCondition(values ...Queryable) Queryable {
    return MultiCondition{
        Values: values,
        Combiner: AndCombiner,
    }
}

func NewMultiListCondition(values ...Queryable) Queryable {
    return MultiCondition{
        Values: values,
        Combiner: CommaCombiner,
    }
}
