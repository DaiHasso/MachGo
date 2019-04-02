package qtypes

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

type testObjectSelectable struct {
    Foo string
}

var _ = Describe("Selectable", func() {
    It("should be createable from a base", func() {
        selectable := BaseSelectable(testObjectSelectable{})

        selectExp, err := selectable()
        Expect(err).ToNot(HaveOccurred())

        tableName, ok := selectExp.Table()
        Expect(ok).To(BeTrue())
        Expect(tableName).To(Equal("test_object_selectables"))
        Expect(selectExp.Column()).To(Equal("*"))
    })
    It("should be createable from a literal with a table name", func() {
        selectable := LiteralSelectable("a.b")

        selectExp, err := selectable()
        Expect(err).ToNot(HaveOccurred())

        tableName, ok := selectExp.Table()
        Expect(ok).To(BeTrue())
        Expect(tableName).To(Equal("a"))
        Expect(selectExp.Column()).To(Equal("b"))
    })
    It("should be createable from a literal without a table name", func() {
        selectable := LiteralSelectable("a")

        selectExp, err := selectable()
        Expect(err).ToNot(HaveOccurred())

        _, ok := selectExp.Table()
        Expect(ok).To(BeFalse())
        Expect(selectExp.Column()).To(Equal("a"))
    })
})
