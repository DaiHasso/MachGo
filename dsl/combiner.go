package dsl

import (
    "fmt"
    "strings"

    "github.com/pkg/errors"
)

type Combiner int

const (
    UnsetCombiner Combiner = iota
    EqualCombiner
    NotEqualCombiner
    GreaterThanCombiner
    GreaterThanEqualCombiner
    LessThanCombiner
    LessThanEqualCombiner
    InCombiner
    AndCombiner
    OrCombiner
    NotCombiner
    CommaCombiner
)

func (self Combiner) String() string {
    switch(self) {
        case EqualCombiner:
        return "="
        case NotEqualCombiner:
        return "!="
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
        case CommaCombiner:
        return ","
    }
    panic(errors.Errorf("Unknown combiner %#+v!", self))
}

func (self Combiner) Join(parts ...string) string {
    var combinerString string

    switch(self) {
        case AndCombiner, OrCombiner:
        combinerString = fmt.Sprintf(" %s ", self.String())
        case CommaCombiner:
        combinerString = fmt.Sprintf("%s ", self.String())
        default:
        combinerString = self.String()
    }
    resultString := strings.Join(parts, combinerString)

    return resultString
}
