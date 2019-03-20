package dsl_test

import (
    "database/sql"
    "fmt"
    "math/rand"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    . "github.com/daihasso/machgo/dsl"
)

var _ = Describe("Queryable", func() {
    Context("When no QuerySequence is provided", func() {
        BeforeEach(func(){
            rand.Seed(1337)
        })
        It("Should create a simple const", func() {
            constInt := ConstantQueryable{[]interface{}{5}}

            stringValue := constInt.String()
            fmt.Fprint(GinkgoWriter, stringValue)

            Expect(stringValue).To(Equal("5"))
        })

        It("Should handle more complex const values", func() {
            foo := struct{foo int}{foo: 5}
            constFoo := ConstantQueryable{[]interface{}{foo}}

            stringValue := constFoo.String()
            fmt.Fprint(GinkgoWriter, stringValue)

            Expect(stringValue).To(Equal("{foo:5}"))
        })

        It("Should handle consts with multiple values", func() {
            constValues := ConstantQueryable{[]interface{}{5, 6, 7, "foo"}}

            stringValue := constValues.String()
            fmt.Fprint(GinkgoWriter, stringValue)

            Expect(stringValue).To(Equal("(5, 6, 7, 'foo')"))
        })

        It("Should handle table columns", func() {
            tableColumn := TableColumnQueryable{
                TableName: "foo",
                ColumnName: "bar",
            }

            stringValue := tableColumn.String()

            fmt.Fprint(GinkgoWriter, stringValue)

            Expect(stringValue).To(Equal("foo.bar"))
        })

        It("Should handle raw columns", func() {
            columnQueryable := ColumnQueryable{
                ColumnName: "bar",
            }

            stringValue := columnQueryable.String()

            fmt.Fprint(GinkgoWriter, stringValue)

            Expect(stringValue).To(Equal("bar"))
        })
    })
    Context("When a QuerySequence is provided", func(){
        var qs *QuerySequence
        BeforeEach(func() {
            qs = NewJoin(object1)
        })

        It("Should handle a simple const", func() {
            constInt := ConstantQueryable{[]interface{}{5}}

            stringValue, args := constInt.QueryValue(qs)
            fmt.Fprint(GinkgoWriter, stringValue)
            fmt.Fprint(GinkgoWriter, args)

            Expect(stringValue).To(Equal("@const_5799089487994996006"))
            expectedArg := sql.Named(stringValue[1:], 5)
            Expect(args).To(Equal([]interface{}{expectedArg}))
        })

        It("Should handle more complex const values", func() {
            foo := struct{foo int}{foo: 5}
            constFoo := ConstantQueryable{[]interface{}{foo}}

            stringValue, args := constFoo.QueryValue(qs)
            fmt.Fprint(GinkgoWriter, stringValue)
            fmt.Fprint(GinkgoWriter, args)

            Expect(stringValue).To(Equal("@const_3156374381228586306"))
            namedArg := sql.Named("const_3156374381228586306", foo)
            Expect(args).To(Equal([]interface{}{namedArg}))
        })


        It("Should handle consts with multiple values", func() {
            values := []interface{}{5, 6, 7, "foo"}
            constValues := ConstantQueryable{values}
            expectedNames := []string{
                "const_3850181338984981652", "const_8857469183970898563",
                "const_8194818716504627408", "const_6133340206446266157",
            }

            stringValue, args := constValues.QueryValue(qs)
            fmt.Fprint(GinkgoWriter, stringValue)
            fmt.Fprint(GinkgoWriter, args)

            Expect(stringValue).To(Equal(
                "(@const_3850181338984981652, @const_8857469183970898563, " +
                    "@const_8194818716504627408, @const_6133340206446266157)",
            ))
            for i, v := range args {
                namedArg := sql.Named(expectedNames[i], values[i])
                Expect(v).To(Equal(namedArg))
            }

        })


        It("Should handle a table & column not in the join", func() {
            tableColumn := TableColumnQueryable{
                TableName: "foo",
                ColumnName: "bar",
            }

            stringValue, args := tableColumn.QueryValue(qs)
            fmt.Fprint(GinkgoWriter, stringValue)
            fmt.Fprint(GinkgoWriter, args)
            Expect(args).Should(BeNil())

            Expect(stringValue).To(MatchRegexp(`foo.bar`))
        })

        It("Should handle a table & column in the join by using it's alias",
            func() {
            tableColumn := TableColumnQueryable{
                TableName: "testtable1",
                ColumnName: "bar",
            }

            stringValue, args := tableColumn.QueryValue(qs)
            fmt.Fprint(GinkgoWriter, stringValue)
            fmt.Fprint(GinkgoWriter, args)
            Expect(args).Should(BeNil())

            Expect(stringValue).To(MatchRegexp(`a.bar`))
        })

        It("Should handle raw columns", func() {
            columnQueryable := ColumnQueryable{
                ColumnName: "bar",
            }

            stringValue, args := columnQueryable.QueryValue(qs)
            fmt.Fprint(GinkgoWriter, stringValue)
            fmt.Fprint(GinkgoWriter, args)
            Expect(args).Should(BeNil())

            Expect(stringValue).To(Equal("bar"))
        })
    })
})
