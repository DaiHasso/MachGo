package dsl

type WhereConditioner func(*[]WhereConditionOrSet)

func addConditionToWhereBuilder(
	conditions *[]WhereConditionOrSet,
	a, b WhereValuer,
	comparison ComparisonOperator,
) {
	whereValuerCondition := WhereValuerCondition{
		Left: a,
		Right: b,
		Comparison: comparison,
		Combiner: UnsetCombiner,
	}
	(*conditions) = append(*conditions, &whereValuerCondition)
}

func WhereEqual(a, b WhereValuer) WhereConditioner {
	resolve := func(conditions *[]WhereConditionOrSet) {
		addConditionToWhereBuilder(conditions, a, b, Equal)
	}
	return resolve
}

func WhereGreater(a, b WhereValuer) WhereConditioner {
	resolve := func(conditions *[]WhereConditionOrSet) {
		addConditionToWhereBuilder(conditions, a, b, GreaterThan)
	}
	return resolve
}

func WhereLess(a, b WhereValuer) WhereConditioner {
	resolve := func(conditions *[]WhereConditionOrSet) {
		addConditionToWhereBuilder(conditions, a, b, LessThan)
	}
	return resolve
}

func WhereGreaterEqual(a, b WhereValuer) WhereConditioner {
	resolve := func(conditions *[]WhereConditionOrSet) {
		addConditionToWhereBuilder(conditions, a, b, GreaterThanEqual)
	}
	return resolve
}

func WhereLessEqual(a, b WhereValuer) WhereConditioner {
	resolve := func(conditions *[]WhereConditionOrSet) {
		addConditionToWhereBuilder(conditions, a, b, LessThanEqual)
	}
	return resolve
}

func WhereIn(a, b WhereValuer) WhereConditioner {
	resolve := func(conditions *[]WhereConditionOrSet) {
		addConditionToWhereBuilder(conditions, a, b, In)
	}
	return resolve
}

func combineConditioners(
	conditioners []WhereConditioner,
	conditions *[]WhereConditionOrSet,
	combiner ConditionCombiner,
) {
	subConditions := make([]WhereConditionOrSet, 0)

	for i, conditioner := range conditioners {
		conditioner(&subConditions)
		subConditions[i].SetCombiner(AndCombiner)
	}

	conditionSet := WhereConditionSet{
		Conditions: subConditions,
		Combiner: UnsetCombiner,
	}

	(*conditions) = append(*conditions, &conditionSet)
}

func AndWhere(conditioners ...WhereConditioner) WhereConditioner {
	resolve := func(conditions *[]WhereConditionOrSet) {
		combineConditioners(conditioners, conditions, AndCombiner)
	}
	return resolve
}

func OrWhere(conditioners ...WhereConditioner) WhereConditioner {
	resolve := func(conditions *[]WhereConditionOrSet) {
		combineConditioners(conditioners, conditions, OrCombiner)
	}
	return resolve
}

// TODO: Add a NotWhere function, might need some special massaging.
