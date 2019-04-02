package qtypes

import (
    "database/sql"
    "fmt"
    "math/rand"
   
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var _ = Describe("Combiner", func() {
    checkCondition := func(symbol string, combiner Combiner) func() {
        return func() {
            if combiner == NotCombiner {
                return
            }
            It(fmt.Sprintf("should handle an '%s' condition", symbol), func() {
                expectedString := fmt.Sprintf("(foo %s 5)", symbol)
                if combiner == InCombiner {
                    expectedString = fmt.Sprintf("(foo %s (5))", symbol)
                }

                columnQueryable := ColumnQueryable{
                    ColumnName: "foo",
                }
                queryable := NewDefaultCondition(
                    columnQueryable, InterfaceToQueryable(5), combiner,
                )

                Expect(queryable.String()).To(Equal(expectedString))
            })
        }
    }
    for symbol, combiner := range allTestCombiners {
        When("stringified", checkCondition(symbol, combiner))
    }

    It("should properly NOT a condition", func() {
        expectedString := "(NOT (foo = 5))"
        columnQueryable := ColumnQueryable{
            ColumnName: "foo",
        }
        condition := NewDefaultCondition(
            columnQueryable, InterfaceToQueryable(5), EqualCombiner,
        )

        notCondition := NotCondition{
            Value: condition,
        }

        Expect(notCondition.String()).To(Equal(expectedString))
    })

    It("should properly handle an AND condition of conditions", func() {
        expectedString := `(foo = 5) AND (bar = 'baz')`
        columnQueryable := ColumnQueryable{
            ColumnName: "foo",
        }
        condition := NewDefaultCondition(
            columnQueryable, InterfaceToQueryable(5), EqualCombiner,
        )
        columnQueryable = ColumnQueryable{
            ColumnName: "bar",
        }
        condition2 := NewDefaultCondition(
            columnQueryable, InterfaceToQueryable("baz"), EqualCombiner,
        )

        multiCondition := NewMultiAndCondition(condition, condition2)

        Expect(multiCondition.String()).To(Equal(expectedString))
    })

    It("should properly handle an OR condition of conditions", func() {
        expectedString := `(foo = 5) OR (bar = 'baz')`
        columnQueryable := ColumnQueryable{
            ColumnName: "foo",
        }
        condition := NewDefaultCondition(
            columnQueryable, InterfaceToQueryable(5), EqualCombiner,
        )
        columnQueryable = ColumnQueryable{
            ColumnName: "bar",
        }
        condition2 := NewDefaultCondition(
            columnQueryable, InterfaceToQueryable("baz"), EqualCombiner,
        )

        multiCondition := NewMultiOrCondition(condition, condition2)

        Expect(multiCondition.String()).To(Equal(expectedString))
    })

    When("an AliasedTable is provided", func() {
        var (
            aliasedTables *AliasedTables
            object = &testObjectQueryable{}
        )
        BeforeEach(func() {
            var err error
            aliasedTables, err = NewAliasedTables(object)
            Expect(err).ToNot(HaveOccurred())
            Expect(aliasedTables).ToNot(BeNil())
            fmt.Fprintf(GinkgoWriter, "AliasedTables: %#+v\n", aliasedTables)
            rand.Seed(84842)
        })
        Describe("DefaultCondition", func() {
            It("should correctly namespace constants", func() {
                expectedVar1Name := "const_3394943692176117195"
                expectedVar2Name := "const_5834364151111774219"
                expectedQueryString := fmt.Sprintf(
                    "(:%s = :%s)", expectedVar1Name, expectedVar2Name,
                )

                condition := NewDefaultCondition(
                    InterfaceToQueryable(2),
                    InterfaceToQueryable(5),
                    EqualCombiner,
                )

                queryString, namedArgs := condition.QueryValue(aliasedTables)
                fmt.Fprint(GinkgoWriter, namedArgs)
                namedArg1 := namedArgs[0].(sql.NamedArg)
                namedArg2 := namedArgs[1].(sql.NamedArg)
                Expect(namedArgs).To(HaveLen(2))
                Expect(queryString).To(Equal(expectedQueryString))
                Expect(namedArg1.Value).To(Equal(2))
                Expect(namedArg2.Value).To(Equal(5))
                Expect(namedArg1.Name).To(Equal(expectedVar1Name))
                Expect(namedArg2.Name).To(Equal(expectedVar2Name))
            })
        })
        Describe("DefaultCondition", func() {
            It("should correctly namespace constants", func() {
                expectedVar1Name := "const_3394943692176117195"
                expectedVar2Name := "const_5834364151111774219"
                expectedQueryString := fmt.Sprintf(
                    "(:%s = :%s)", expectedVar1Name, expectedVar2Name,
                )

                condition := NewDefaultCondition(
                    InterfaceToQueryable(2),
                    InterfaceToQueryable(5),
                    EqualCombiner,
                )

                queryString, namedArgs := condition.QueryValue(aliasedTables)
                fmt.Fprint(GinkgoWriter, namedArgs)
                namedArg1 := namedArgs[0].(sql.NamedArg)
                namedArg2 := namedArgs[1].(sql.NamedArg)
                Expect(namedArgs).To(HaveLen(2))
                Expect(queryString).To(Equal(expectedQueryString))
                Expect(namedArg1.Value).To(Equal(2))
                Expect(namedArg2.Value).To(Equal(5))
                Expect(namedArg1.Name).To(Equal(expectedVar1Name))
                Expect(namedArg2.Name).To(Equal(expectedVar2Name))
            })
        })
        Describe("MultiCondition", func() {
            It("should properly namespace an OR condition of conditions",
                func() {
                    var1Name := "const_3394943692176117195"
                    var2Name := "const_5834364151111774219"
                    expectedString := fmt.Sprintf(
                        `(foo = :%s) OR (bar = :%s)`, var1Name, var2Name,
                    )
                    columnQueryable := ColumnQueryable{
                        ColumnName: "foo",
                    }
                    condition := NewDefaultCondition(
                        columnQueryable,
                        InterfaceToQueryable(5),
                        EqualCombiner,
                    )
                    columnQueryable = ColumnQueryable{
                        ColumnName: "bar",
                    }
                    condition2 := NewDefaultCondition(
                        columnQueryable,
                        InterfaceToQueryable("baz"),
                        EqualCombiner,
                    )

                    multiCondition := NewMultiOrCondition(
                        condition, condition2,
                    )

                    queryString, args := multiCondition.QueryValue(
                        aliasedTables,
                    )
                    fmt.Fprint(GinkgoWriter, args)
                    namedArg1 := args[0].(sql.NamedArg)
                    namedArg2 := args[1].(sql.NamedArg)
                    Expect(queryString).To(Equal(expectedString))
                    Expect(namedArg1.Name).To(Equal(var1Name))
                    Expect(namedArg2.Name).To(Equal(var2Name))
                    Expect(namedArg1.Value).To(Equal(5))
                    Expect(namedArg2.Value).To(Equal("baz"))
                },
            )
        })
    })
})
