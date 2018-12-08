package dsl

import (
	"database/sql"
	"fmt"

	dsl1 "MachGo/dsl"
)

type ConditionEvaluator func(
	*dsl1.QuerySequence, Queryable, Queryable,
) (string, []sql.NamedArg)


func NewDefaultEvaluator(
	combiner Combiner, valueModifiers ...ValueModifier,
) ConditionEvaluator {
	return func(
		qs *dsl1.QuerySequence,
		lhValues,
		rhValues Queryable,
	) (string, []sql.NamedArg) {
		var allArgs []sql.NamedArg
		var leftQueryString, rightQueryString string

		if qs != nil {
			var leftArgs, rightArgs []sql.NamedArg
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
	qs *dsl1.QuerySequence,
) (string, []sql.NamedArg) {
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
