package qtypes

import (
    "fmt"
    "math/rand"
   
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

type testObjectQueryOptions struct {
    Foo string `db:"foo"`
}

var _ = Describe("QueryOption", func() {
    When("an AliasedTable is provided", func() {
        var (
            aliasedTables *AliasedTables
            object = &testObjectQueryOptions{}
        )
        BeforeEach(func() {
            var err error
            aliasedTables, err = NewAliasedTables(object)
            Expect(err).ToNot(HaveOccurred())
            Expect(aliasedTables).ToNot(BeNil())
            fmt.Fprintf(GinkgoWriter, "%#+v", aliasedTables)
            rand.Seed(84842)
        })
        Describe("LimitOption", func() {
            It("should create a proper limit clause", func() {
                expectedQueryString := "LIMIT 10"
                q := LimitOption{
                    Limit: 10,
                }

                queryString, args := q.QueryValue(aliasedTables)

                Expect(queryString).To(Equal(expectedQueryString))
                Expect(args).To(BeEmpty())
            })
        })
        Describe("OffsetOption", func() {
            It("should create a proper offset clause", func() {
                expectedQueryString := "OFFSET 10"
                q := OffsetOption{
                    Offset: 10,
                }

                queryString, args := q.QueryValue(aliasedTables)

                Expect(queryString).To(Equal(expectedQueryString))
                Expect(args).To(BeEmpty())
            })
        })
        Describe("OrderByOption", func() {
            It("should create a proper order by clause", func() {
                expectedQueryString := "ORDER BY a.foo"
                q := OrderByOption{
                    Order: TableColumnQueryable{
                        TableName: "test_object_query_options",
                        ColumnName: "foo",
                    },
                }

                queryString, args := q.QueryValue(aliasedTables)

                Expect(queryString).To(Equal(expectedQueryString))
                Expect(args).To(BeEmpty())
            })
        })
    })
    When("an AliasedTable is not provided", func() {
        Describe("LimitOption", func() {
            It("should create a proper limit clause", func() {
                expectedQueryString := "LIMIT 10"
                q := LimitOption{
                    Limit: 10,
                }

                queryString := q.String()

                Expect(queryString).To(Equal(expectedQueryString))
            })
        })
        Describe("OffsetOption", func() {
            It("should create a proper offset clause", func() {
                expectedQueryString := "OFFSET 10"
                q := OffsetOption{
                    Offset: 10,
                }

                queryString := q.String()

                Expect(queryString).To(Equal(expectedQueryString))
            })
        })
        Describe("OrderByOption", func() {
            It("should create a proper order by clause", func() {
                expectedQueryString := "ORDER BY test_object_query_options.foo"
                q := OrderByOption{
                    Order: TableColumnQueryable{
                        TableName: "test_object_query_options",
                        ColumnName: "foo",
                    },
                }

                queryString := q.String()

                Expect(queryString).To(Equal(expectedQueryString))
            })
        })
    })
})
