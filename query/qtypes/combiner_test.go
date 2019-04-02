package qtypes

import (
    "fmt"
   
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var allTestCombiners = map[string]Combiner {
    "=": EqualCombiner,
    "!=": NotEqualCombiner,
    ">": GreaterThanCombiner,
    ">=": GreaterThanEqualCombiner,
    "<": LessThanCombiner,
    "<=": LessThanEqualCombiner,
    "IN": InCombiner,
    "AND": AndCombiner,
    "OR": OrCombiner,
    "NOT": NotCombiner,
    ",": CommaCombiner,
}

var _ = Describe("Combiner", func() {
    checkCombiner := func(symbol string, combiner Combiner) func() {
        return func() {
            It(
                fmt.Sprintf(
                    "should generate a string for a '%s' condition",
                    symbol,
                ),
                func() {
                    combinerString := combiner.String()
                    Expect(combinerString).To(Equal(symbol))
                },
            )
        }
    }

    checkJoin := func(symbol string, combiner Combiner) func() {
        return func() {
            It(
                fmt.Sprintf(
                    "should generate a join string for a '%s' condition",
                    symbol,
                ),
                func() {
                    arg1, arg2 := "5", "6"
                    combinerString := combiner.Join(arg1, arg2)
                    expectedString := fmt.Sprintf("%s%s%s", arg1, symbol, arg2)
                    if combiner == OrCombiner || combiner == AndCombiner ||
                        combiner == NotCombiner {
                        expectedString = fmt.Sprintf(
                            "%s %s %s", arg1, symbol, arg2,
                        )
                    } else if combiner == CommaCombiner {
                        expectedString = fmt.Sprintf(
                            "%s%s %s", arg1, combiner, arg2,
                        )
                    }
                    Expect(combinerString).To(Equal(expectedString))
                },
            )
        }
    }

    for symbol, combiner := range allTestCombiners {
        When("stringified", checkCombiner(symbol, combiner))
    }

    for symbol, combiner := range allTestCombiners {
        When("joined", checkJoin(symbol, combiner))
    }
})
