package dsl

import (
	"fmt"
	"reflect"
	"unicode"
	"database/sql/driver"

	"github.com/DaiHasso/MachGo"
	"github.com/DaiHasso/MachGo/refl"
)

// WhereValuer is a function that takes a QuerySequence and returns a
// WhereValue resolved using that QuerySequence.
type WhereValuer (func(*QuerySequence) WhereValue)

func Const(values ...driver.Value) WhereValuer {
	isMultiple := false
	if len(values) > 1 {
		isMultiple = true
	}
	resolve := func(qs *QuerySequence) WhereValue {
		return ConstantValue{
			values: values,
			isMultiple: isMultiple,
		}
	}

	return resolve
}

func ObjectColumn2(obj MachGo.Object, column string) WhereValuer {
	resolve := func (qs *QuerySequence) WhereValue {
		if qs == nil {
			// This handles printing of the query.
			return NamespacedColumn{
				isNamespaced: true,
				columnName: column,
				tableAlias: "",
				tableNamespace: obj.GetTableName(),
			}
		}

		found := false
		columnName := ""
		objType := refl.Deref(reflect.TypeOf(obj))
		if unicode.IsUpper(rune(column[0])) {
			fieldNameBSFields := *qs.typeFieldNameBSFieldMap[objType]
			if value, ok := fieldNameBSFields[column]; ok {
				columnName = value.Tag("db").Value()
				found = true
			}
		}
		if !found {
			_, ok := qs.typeBSFieldMap[objType]
			if !ok {
				panic(fmt.Errorf(
					"Column name '%s' is neither a name of a property or a " +
						"name of a db tag in the QuerySequence.",
					column,
				))
			}

			columnName = column
		}

		tableAlias, ok := qs.tableAliasMap[obj.GetTableName()]
		if !ok {
			panic(fmt.Errorf(
				"Object '%s' not in QuerySequence!",
				refl.GetInterfaceName(obj),
			))
		}

		namespacedColumn := NamespacedColumn{
			isNamespaced: true,
			columnName: columnName,
			tableAlias: tableAlias,
			tableNamespace: obj.GetTableName(),
		}

		return namespacedColumn
	}

	return resolve
}

type WhereValuerCondition struct {
	Left,
	Right WhereValuer
	Comparison ComparisonOperator
	Combiner ConditionCombiner
}

func (self *WhereValuerCondition) GetCombiner() ConditionCombiner {
    return self.Combiner
}

func (self *WhereValuerCondition) SetCombiner(combiner ConditionCombiner) {
    self.Combiner = combiner
}

func (self *WhereValuerCondition) String() string {
    leftString := conditionValueString(self.Left(nil))
    rightString := conditionValueString(self.Right(nil))
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
func (self *WhereValuerCondition) ToQuery(
	qs *QuerySequence,
) (string, []interface{}) {
    var result string
    var values []interface{}

    leftString, leftValues := self.Left(qs).QueryValue(qs)
    if leftValues != nil {
        values = append(values, leftValues...)
    }
    result += leftString

    result += fmt.Sprintf(" %s ", comparisonToString(self.Comparison))

    rightString, rightValues := self.Right(qs).QueryValue(qs)
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
