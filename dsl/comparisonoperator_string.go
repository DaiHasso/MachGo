// Code generated by "stringer -type=ComparisonOperator"; DO NOT EDIT.

package dsl

import "strconv"

const _ComparisonOperator_name = "UnsetComparisonOperatorEqualGreaterThanLessThanGreaterThanEqualLessThanEqualIn"

var _ComparisonOperator_index = [...]uint8{0, 23, 28, 39, 47, 63, 76, 78}

func (i ComparisonOperator) String() string {
	if i < 0 || i >= ComparisonOperator(len(_ComparisonOperator_index)-1) {
		return "ComparisonOperator(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ComparisonOperator_name[_ComparisonOperator_index[i]:_ComparisonOperator_index[i+1]]
}
