package dsl

import (
  "github.com/pkg/errors"
)

type Combiner int

const (
    UnsetCombiner Combiner = iota
    EqualCombiner
    GreaterThanCombiner
    GreaterThanEqualCombiner
    LessThanCombiner
    LessThanEqualCombiner
    InCombiner
	AndCombiner
	OrCombiner
	NotCombiner
)

func (self Combiner) String() string {
	switch(self) {
		case EqualCombiner:
		return "="
		case GreaterThanCombiner:
		return ">"
		case GreaterThanEqualCombiner:
		return ">="
		case LessThanCombiner:
		return "<"
		case LessThanEqualCombiner:
		return "<="
		case InCombiner:
		return "IN"
		case AndCombiner:
		return "AND"
		case OrCombiner:
		return "OR"
		case NotCombiner:
		return "NOT"
	}
	panic(errors.Errorf("Unknown combiner %s!", self))
}
