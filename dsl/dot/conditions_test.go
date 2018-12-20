package dot_test


import (
	"fmt"
	"math/rand"

    . "github.com/onsi/ginkgo"
    t "github.com/onsi/gomega"
    . "MachGo/dsl/dot"
    "MachGo/dsl"
)

var _ = Describe("Conditions", func() {
	type conditionFunc func(interface{}, interface{}) dsl.Queryable
	conditions := map[string]conditionFunc {
		"=": Equal,
		"!=": NotEqual,
		">": GreaterThan,
		"<": LessThan,
		">=": GreaterThanEqual,
		"<=": LessThanEqual,
	}
	conditionAliases := map[string]conditionFunc {
		"=": Eq,
		"!=": Neq,
		">": Gt,
		"<": Lt,
		">=": Gte,
		"<=": Lte,
	}
	When("No query sequence is provided", func() {
		checkCondition := func(
			name, symbol string, fn conditionFunc,
		) func() {
			return func() {
				It(fmt.Sprintf(
					"Should generate a proper %s where condition when raw " +
						"ints are provided",
					symbol,
				), func() {
					// #nosec G404
					a, b := rand.Int(), rand.Int()
					condition := fn(a, b)
					conditionString := condition.String()

					fmt.Fprint(GinkgoWriter, conditionString)

					expectedString := fmt.Sprintf("(%d %s %d)", a, symbol, b)
					t.Expect(conditionString).To(t.Equal(expectedString))
				})

				It(fmt.Sprintf(
					"Should generate a proper %s where condition when " +
						"queryables are provided",
					symbol,
				), func() {
					// #nosec G404
					a, b := Const(rand.Int()), Const(rand.Int())
					condition := fn(a, b)
					conditionString := condition.String()

					fmt.Fprint(GinkgoWriter, conditionString)

					expectedString := fmt.Sprintf(
						"(%s %s %s)", a.String(), symbol, b.String(),
					)
					t.Expect(conditionString).To(t.Equal(expectedString))
				})
			}
		}
		for symbol, fn := range conditions {
			blurb := fmt.Sprintf(
				"When a %s condition is used", symbol,
			)
			Context(blurb, checkCondition(symbol, symbol, fn))
		}
		for symbol, fn := range conditionAliases {
			blurb := fmt.Sprintf(
				"When a %s condition alias is used",
				symbol,
			)
			Context(blurb, checkCondition(symbol, symbol, fn))
		}

		It("Should generate a proper In where condition", func() {
			lhs := ObjectColumn(object1, "foo")

			condition := In(lhs, 1, 2, 3)
			conditionString := condition.String()

			fmt.Fprint(GinkgoWriter, conditionString)

			expectedString := fmt.Sprintf("(testtable1.foo IN (1, 2, 3))")

			t.Expect(conditionString).To(t.Equal(expectedString))
		})

		It("Should generate a proper And where condition", func() {
			condition := And(
				Eq(ObjectColumn(object1, "foo"), 1),
				In(ObjectColumn(object2, "bar"), 2, 3, 4),
			)

			conditionString :=  condition.String()

			fmt.Fprint(GinkgoWriter, conditionString)

			expectedString := fmt.Sprintf(
				"(testtable1.foo = 1) AND (testtable2.bar IN (2, 3, 4))",
			)

			t.Expect(conditionString).To(t.Equal(expectedString))
		})

		It("Should generate a proper Or where condition", func() {
			condition := Or(
				Eq(ObjectColumn(object1, "foo"), 1),
				In(ObjectColumn(object2, "bar"), 2, 3, 4),
			)

			conditionString :=  condition.String()

			fmt.Fprint(GinkgoWriter, conditionString)

			expectedString := fmt.Sprintf(
				"(testtable1.foo = 1) OR (testtable2.bar IN (2, 3, 4))",
			)

			t.Expect(conditionString).To(t.Equal(expectedString))
		})

		It("Should generate a proper Not where condition", func() {
			condition := And(
				Not(Eq(ObjectColumn(object1, "foo"), 1)),
				In(ObjectColumn(object2, "bar"), 2, 3, 4),
			)

			conditionString :=  condition.String()

			fmt.Fprint(GinkgoWriter, conditionString)

			expectedString := fmt.Sprintf(
				"(NOT (testtable1.foo = 1)) AND (testtable2.bar IN (2, 3, 4))",
			)

			t.Expect(conditionString).To(t.Equal(expectedString))
		})
	})
})
