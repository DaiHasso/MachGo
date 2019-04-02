package qtypes

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

var _ = Describe("SelectExpression", func() {
    It("should create handle a column with a table", func() {
        selectExp := NewSelectExpression("a.b")

        tableName, ok := selectExp.Table()
        Expect(ok).To(BeTrue())
        Expect(selectExp.Column()).To(Equal("b"))
        Expect(tableName).To(Equal("a"))
    })
    It("should create handle a column without a table", func() {
        selectExp := NewSelectExpression("b")

        _, ok := selectExp.Table()
        Expect(ok).To(BeFalse())
        Expect(selectExp.Column()).To(Equal("b"))
    })
})
