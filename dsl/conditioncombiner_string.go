// Code generated by "stringer -type=ConditionCombiner"; DO NOT EDIT.

package dsl

import "strconv"

const _ConditionCombiner_name = "UnsetCombinerAndCombinerOrCombiner"

var _ConditionCombiner_index = [...]uint8{0, 13, 24, 34}

func (i ConditionCombiner) String() string {
	if i < 0 || i >= ConditionCombiner(len(_ConditionCombiner_index)-1) {
		return "ConditionCombiner(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ConditionCombiner_name[_ConditionCombiner_index[i]:_ConditionCombiner_index[i+1]]
}