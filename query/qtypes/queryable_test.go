package qtypes

import (
    "database/sql"
    "fmt"
    "math/rand"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

type testObjectQueryable struct {
    Foo string
}

var _ = Describe("Queryable", func() {
    Describe("InterfaceToQueryable", func() {
        It("should create a ConstantQueryable from an interface", func() {
            queryable := InterfaceToQueryable(5)
            Expect(queryable).To(BeAssignableToTypeOf(ConstantQueryable{}))
        })
        It("should return the passed interface if it is already a queryable",
            func() {
                queryable := InterfaceToQueryable(5)
                secondQueryable := InterfaceToQueryable(queryable)
                Expect(secondQueryable).To(BeEquivalentTo(queryable))
            },
        )
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
            fmt.Fprintf(GinkgoWriter, "%#+v", aliasedTables)
            rand.Seed(84842)
        })
        Describe("ConstantQueryable", func() {
            It("should provide collision-free value", func() {
                expectedQueryString := "@const_3394943692176117195"
                q := InterfaceToQueryable(5)

                queryString, namedArgs := q.QueryValue(aliasedTables)

                namedArg := namedArgs[0].(sql.NamedArg)

                Expect(queryString).To(Equal(expectedQueryString))
                Expect(namedArg.Value).To(Equal(5))
                Expect(namedArg.Name).To(Equal(expectedQueryString[1:]))
            })
            It("should provide collision-free values when provided multiple " +
                "values", func() {
                expectedName := "const_3394943692176117195"
                expectedName2 := "const_5834364151111774219"
                expectedQueryString := fmt.Sprintf(
                    "(@%s, @%s)", expectedName, expectedName2,
                )
                q := ConstantQueryable{
                    Values: []interface{}{5, 6},
                }

                queryString, namedArgs := q.QueryValue(aliasedTables)
                Expect(namedArgs).To(HaveLen(2))
                fmt.Fprintf(GinkgoWriter, queryString)
                Expect(queryString).To(Equal(expectedQueryString))

                namedArg := namedArgs[0].(sql.NamedArg)
                namedArg2 := namedArgs[1].(sql.NamedArg)
                Expect(namedArg.Value).To(Equal(5))
                Expect(namedArg.Name).To(Equal(expectedName))
                Expect(namedArg2.Value).To(Equal(6))
                Expect(namedArg2.Name).To(Equal(expectedName2))
            })
        })
        Describe("TableColumnQueryable", func() {
            It("should grab the appropriate alias", func() {
                tableColumnQueryable := TableColumnQueryable{
                    TableName: "test_object_queryables",
                    ColumnName: "foo",
                }
                queryString, namedArgs := tableColumnQueryable.QueryValue(
                    aliasedTables,
                )

                Expect(namedArgs).To(BeNil())
                Expect(queryString).To(Equal("a.foo"))
            })
        })
        Describe("ColumnQueryable", func() {
            It("should print the column directly", func() {
                tableColumnQueryable := ColumnQueryable{
                    ColumnName: "foo",
                }
                queryString, args := tableColumnQueryable.QueryValue(
                    aliasedTables,
                )
                Expect(args).To(BeNil())

                Expect(queryString).To(Equal("foo"))
            })
        })
    })
    When("an AliasedTable is not provided", func() {
        Describe("ConstantQueryable", func() {
            It("should provide the actual value", func() {
                expectedQueryString := "5"
                q := InterfaceToQueryable(5)

                queryString := q.String()

                Expect(queryString).To(Equal(expectedQueryString))
            })
            It("should provide properly formated values when multiple are " +
                "provided", func() {
                expectedName := "5"
                expectedName2 := "'foo'"
                expectedQueryString := fmt.Sprintf(
                    "(%s, %s)", expectedName, expectedName2,
                )
                q := ConstantQueryable{
                    Values: []interface{}{5, "foo"},
                }

                queryString := q.String()
                fmt.Fprintf(GinkgoWriter, queryString)
                Expect(queryString).To(Equal(expectedQueryString))
            })
        })
        Describe("TableColumnQueryable", func() {
            It("should grab directly use the table name", func() {
                tableColumnQueryable := TableColumnQueryable{
                    TableName: "test_object_queryables",
                    ColumnName: "foo",
                }
                queryString := tableColumnQueryable.String()

                Expect(queryString).To(Equal("test_object_queryables.foo"))
            })
        })
        Describe("ColumnQueryable", func() {
            It("should print the column directly", func() {
                tableColumnQueryable := ColumnQueryable{
                    ColumnName: "foo",
                }
                queryString := tableColumnQueryable.String()

                Expect(queryString).To(Equal("foo"))
            })
        })
    })
    Describe("ObjectColumn", func() {
        var object = &testObjectQueryable{}
        It("should should create a TableColumnQueryable with the correct " +
            "table and column", func() {
            q, err := ObjectColumn(object, "foo")
            Expect(err).ToNot(HaveOccurred())

            tcq := q.(TableColumnQueryable)
            Expect(tcq.TableName).To(Equal("test_object_queryables"))
        })
    })
})
