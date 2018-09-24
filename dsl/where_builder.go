package dsl

import (
    "database/sql/driver"
    "fmt"
    "reflect"
    "unicode"

    database "github.com/DaiHasso/MachGo"

    logging "github.com/daihasso/slogging"
)

// ComparisonOperator is a type of comparison for a where condition.
//go:generate stringer -type=ComparisonOperator
type ComparisonOperator int

// These are all the comparison operators for a where statement.
const (
    UnsetComparisonOperator ComparisonOperator = iota
    Equal
    GreaterThan
    LessThan
    GreaterThanEqual
    LessThanEqual
    In
)

// ConditionCombiner is a combiner for conditions in a where clause.
//go:generate stringer -type=ConditionCombiner
type ConditionCombiner int

// There are the combiners for conditions in where clauses
const (
    UnsetCombiner ConditionCombiner = iota
    AndCombiner
    OrCombiner
)

// WhereBuilder creates a set of where conditions.
// TODO: Handle parenthesised queries.
type WhereBuilder struct {
    CurrentCondition *WhereCondition
    Conditions       []WhereConditionOrSet
}

// WhereConditionOrSet represents either a single condition or a collection of
// conditions in as a sub (parenthesised) condition.
type WhereConditionOrSet interface {
    fmt.Stringer

    ToQuery(*QuerySequence) (string, []interface{})
    GetCombiner() ConditionCombiner
    SetCombiner(ConditionCombiner)
}

// WhereCondition is a condition in a where clause.
// TODO: This should probably be an implementer of an interface that supports
//       sub-expresions.
type WhereCondition struct {
    Left,
    Right WhereValue
    Comparison ComparisonOperator
    Combiner   ConditionCombiner
}

func (self *WhereCondition) GetCombiner() ConditionCombiner {
    return self.Combiner
}

func (self *WhereCondition) SetCombiner(combiner ConditionCombiner) {
    self.Combiner = combiner
}

func (self *WhereCondition) String() string {
    leftString := conditionValueString(self.Left)
    rightString := conditionValueString(self.Right)
    operatorString := comparisonToString(self.Comparison)

    conditionString := fmt.Sprintf(
        "%s %s %s",
        leftString,
        operatorString,
        rightString,
    )

    return conditionString
}

// ToQuery takes a QuerySequence and converts itself to a database-ready query
// string and values for bindvars.
func (self *WhereCondition) ToQuery(qs *QuerySequence) (string, []interface{}) {
    var result string
    var values []interface{}

    leftString, leftValues := self.Left.QueryValue(qs)
    if leftValues != nil {
        values = append(values, leftValues...)
    }
    result += leftString

    result += fmt.Sprintf(" %s ", comparisonToString(self.Comparison))

    rightString, rightValues := self.Right.QueryValue(qs)
    if rightValues != nil {
        values = append(values, rightValues...)
    }
    if self.Comparison == In {
        result += "(" + rightString + ")"
    } else {
        result += rightString
    }

    return result, values
}

type WhereConditionSet struct {
    Conditions []WhereConditionOrSet
    Combiner   ConditionCombiner
}

func (self *WhereConditionSet) GetCombiner() ConditionCombiner {
    return self.Combiner
}

func (self *WhereConditionSet) SetCombiner(combiner ConditionCombiner) {
    self.Combiner = combiner
}

func (self *WhereConditionSet) String() string {
    result := "("
    for _, conditionOrSet := range self.Conditions {
        if len(result) > 1 {
            result += " "
        }

        conditionString := conditionOrSet.String()
        result += conditionString

        combiner := conditionOrSet.GetCombiner()
        if combiner != UnsetCombiner {
            result += " " + combinerToString(combiner)
        }
    }

    result += ")"

    return result
}

// ToQuery takes a QuerySequence and converts itself to a database-ready query
// string and values for bindvars.
func (self *WhereConditionSet) ToQuery(qs *QuerySequence) (string, []interface{}) {
    result := "("
    var values []interface{}
    for _, conditionOrSet := range self.Conditions {
        if len(result) > 1 {
            result += " "
        }

        conditionString, conditionValues := conditionOrSet.ToQuery(qs)
        combinerString := combinerToString(conditionOrSet.GetCombiner())

        result += fmt.Sprintf("%s %s", conditionString, combinerString)

        values = append(values, conditionValues...)
    }

    result += ")"

    return result, values
}

func combinerToString(combiner ConditionCombiner) string {
    switch combiner {
    case AndCombiner:
        return "AND"
    case OrCombiner:
        return "OR"
    default:
        panic(fmt.Sprintf(
            "Unknown combination operator: '%s'",
            fmt.Sprint(combiner),
        ))
    }
}

func comparisonToString(operator ComparisonOperator) string {
    switch operator {
    case Equal:
        return "="
    case GreaterThan:
        return ">"
    case LessThan:
        return "<"
    case GreaterThanEqual:
        return ">="
    case LessThanEqual:
        return "<="
    case In:
        return "in"
    default:
        panic(fmt.Sprintf(
            "Unknown comparison operator: '%s'",
            fmt.Sprint(operator),
        ))
    }
}

func NewWhere() *WhereBuilder {
    return new(WhereBuilder)
}

func (self *WhereBuilder) ObjectColumn(
    obj database.Object,
    column string,
) *WhereBuilder {
    columnString := column
    if unicode.IsUpper(rune(column[0])) {
        field, found := reflect.TypeOf(obj).FieldByName(column)
        if found {
            if tagValue, ok := field.Tag.Lookup(`db`); ok {
                columnString = tagValue
            }
        }
    }
    namespacedColumn := NamespacedColumn{
        true,
        columnString,
        "",
        obj.GetTableName(),
    }

    self.setCurrentConditionValue(namespacedColumn)

    return self
}

func (self *WhereBuilder) Eq(values ...WhereValue) *WhereBuilder {
    self.checkAndSetComparison(Equal)

    if len(values) == 1 {
        self.setCurrentConditionValue(values[0])
    }
    // TODO: Handle multiple values, maybe duplicate and AND conditions with all
    //       the values?

    return self
}

func (self *WhereBuilder) Greater(values ...WhereValue) *WhereBuilder {
    self.checkAndSetComparison(GreaterThan)

	return self
}

func (self *WhereBuilder) Less(values ...WhereValue) *WhereBuilder {
	self.checkAndSetComparison(LessThan)

	return self
}

func (self *WhereBuilder) GreaterEq(values ...WhereValue) *WhereBuilder {
	self.checkAndSetComparison(GreaterThanEqual)

	return self
}

func (self *WhereBuilder) LessEq(values ...WhereValue) *WhereBuilder {
	self.checkAndSetComparison(LessThanEqual)

	return self
}

func (self *WhereBuilder) In(values ...WhereValue) *WhereBuilder {
	self.checkAndSetComparison(In)

	return self
}

func (self *WhereBuilder) Const(values ...driver.Value) *WhereBuilder {
	isMultiple := false
	if len(values) > 1 {
		isMultiple = true
	}
	whereValue := ConstantValue{
		values: values,
		isMultiple: isMultiple,
	}

	self.setCurrentConditionValue(whereValue)

	return self
}

func (self *WhereBuilder) And() *WhereBuilder {
	self.checkAndSetCombiner(AndCombiner)

	return self
}

func (self *WhereBuilder) Or() *WhereBuilder {
	self.checkAndSetCombiner(OrCombiner)

	return self
}

func (self *WhereBuilder) SubCond(
	otherWhereBuilder *WhereBuilder,
) *WhereBuilder {
	// TODO: Checking of in-progress current condition.

	self.Conditions = append(
		self.Conditions,
		otherWhereBuilder.asSubCondition(),
	)

	return self
}

func (self *WhereBuilder) String() string {
	whereString, _ := self.buildWhere(nil)
	return whereString
}

func (self *WhereBuilder) asQuery(qs *QuerySequence) (string, []interface{}) {
	return self.buildWhere(qs)
}

func (self *WhereBuilder) asSubCondition() WhereConditionOrSet {
	newSet := &WhereConditionSet{
		Conditions: self.Conditions,
		Combiner:   UnsetCombiner,
	}

	return newSet
}

func (self *WhereBuilder) buildWhere(qs *QuerySequence) (
	string,
	[]interface{},
) {
	var result string
	var values []interface{}
	for i, condition := range self.Conditions {
		if len(result) > 0 {
			result += " "
		}
		if qs == nil {
			result += condition.String()
		} else {
			conditionString, conditionValues := condition.ToQuery(qs)
			result += conditionString
			values = append(values, conditionValues...)
		}
		if i == (len(self.Conditions) - 1) {
			// Don't add a combiner to the last element.
			continue
		}
		combiner := condition.GetCombiner()
		if combiner != UnsetCombiner {
			combinerString := combinerToString(combiner)
			result += " " + combinerString
		} else if i != len(self.Conditions) - 1 {
			// Default to AND if nothing's set, this behavior is debatable but
			// seems to be pretty standard among ORMs.
			combinerString := combinerToString(AndCombiner)
			result += " " + combinerString
		}
	}

	return result, values
}

func conditionValueString(
	value WhereValue,
) string {
	val, err := value.Value()
	if err != nil {
		// TODO: Is this a panic?
		panic(err)
	}

	return fmt.Sprint(val)
}

func (self *WhereBuilder) checkAndSetComparison(comparison ComparisonOperator) {
	if self.CurrentCondition == nil {
		panic(fmt.Sprintf(
			"Can't use %s without specifying a lefthand variable.",
			comparison,
		))
	}

	if self.CurrentCondition.Comparison != UnsetComparisonOperator {
		panic("Can't set more than one condition operator.")
	}

	self.CurrentCondition.Comparison = comparison
}

func (self *WhereBuilder) checkAndSetCombiner(combiner ConditionCombiner) {
	if self.CurrentCondition == nil {
		if len(self.Conditions) == 0 {
			logging.Warn(
				"Used combiner in a where clause with no conditions.",
			).With("combiner", combiner).Send()
		}
		lastCondition := self.Conditions[len(self.Conditions)-1]

		if lastCondition.GetCombiner() != UnsetCombiner {
			logging.Warn(
				"Tried to set condition combination when a combination already "+
					"was set.",
			).With(
				"existing_combiner",
				lastCondition.GetCombiner(),
			).With(
				"new_combiner",
				combiner,
			).Send()
		}

		lastCondition.SetCombiner(combiner)
	} else {
		panic("Can't combine conditions without finishing current condition.")
	}
}

func (self *WhereBuilder) setCurrentConditionValue(value WhereValue) {
	if self.CurrentCondition != nil {
		if self.CurrentCondition.Right != nil {
			panic(fmt.Sprintf(
				"Condition already has a right and left value: (%s, %s).",
				self.CurrentCondition.Left,
				self.CurrentCondition.Right,
			))
		}
		self.CurrentCondition.Right = value

		self.Conditions = append(self.Conditions, self.CurrentCondition)
		self.CurrentCondition = nil
	} else {
		self.CurrentCondition = &WhereCondition{
			Left:       value,
			Right:      nil,
			Comparison: UnsetComparisonOperator,
			Combiner:   UnsetCombiner,
		}
	}
}
