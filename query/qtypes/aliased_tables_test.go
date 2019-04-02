package qtypes

import (
    "fmt"
    "reflect"
   
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
)

type testObjectAT struct {}
type testOtherObjectAT struct {}

var _ = Describe("AliasedTables", func() {
    var (
        err error
        aliasedTables *AliasedTables
        object = &testObjectAT{}
        object2 = &testOtherObjectAT{}
    )

    BeforeEach(func() {
        aliasedTables, err = NewAliasedTables(object, object2)
        Expect(err).ToNot(HaveOccurred())
        fmt.Fprint(GinkgoWriter, aliasedTables)
        Expect(aliasedTables).ToNot(BeNil())
    })

    It("should create unique aliases for each object", func() {
        aliases := aliasedTables.Aliases()
        Expect(aliases).To(ConsistOf("a", "b"))
    })
    It("should be able to get types for aliases", func() {
        expectedType := reflect.TypeOf(*object)
        expectedType2 := reflect.TypeOf(*object2)

        Expect(*aliasedTables.TypeForAlias("a")).To(Equal(expectedType))
        Expect(*aliasedTables.TypeForAlias("b")).To(Equal(expectedType2))
    })
    It("should be able to get types for table names", func() {
        expectedType := reflect.TypeOf(*object)
        expectedType2 := reflect.TypeOf(*object2)

        Expect(
            *aliasedTables.TypeForTable("test_object_ats"),
        ).To(Equal(expectedType))
        Expect(
            *aliasedTables.TypeForTable("test_other_object_ats"),
        ).To(Equal(expectedType2))
    })
    It("should be able to get table names for aliases", func() {
        tableName1 := aliasedTables.TableForAlias("a")
        tableName2 := aliasedTables.TableForAlias("b")
        Expect(tableName1).To(Equal("test_object_ats"))
        Expect(tableName2).To(Equal("test_other_object_ats"))
    })
    It("should be able to get aliases for table names", func() {
        alias1, ok := aliasedTables.AliasForTable("test_object_ats")
        Expect(ok).To(BeTrue())
        alias2, ok := aliasedTables.AliasForTable("test_other_object_ats")
        Expect(ok).To(BeTrue())
        Expect(alias1).To(Equal("a"))
        Expect(alias2).To(Equal("b"))
    })
    It("should be able to check if object is aliased", func() {
        Expect(aliasedTables.ObjectIsAliased(object)).To(BeTrue())
        Expect(aliasedTables.ObjectIsAliased(object2)).To(BeTrue())
    })
    It("should be able to get alias for object", func() {
        alias, err := aliasedTables.ObjectAlias(object)
        Expect(err).ToNot(HaveOccurred())
        alias2, err := aliasedTables.ObjectAlias(object2)
        Expect(err).ToNot(HaveOccurred())
        Expect(alias).To(Equal("a"))
        Expect(alias2).To(Equal("b"))
    })
    It("should be able to get table for type", func() {
        type1 := reflect.TypeOf(*object)
        type2 := reflect.TypeOf(*object2)

        table := aliasedTables.TypeTable(type1)
        table2 := aliasedTables.TypeTable(type2)
        Expect(table).To(Equal("test_object_ats"))
        Expect(table2).To(Equal("test_other_object_ats"))
    })
    It("should be able to add an object", func() {
        aliasedTables, err = NewAliasedTables()
        Expect(aliasedTables.Aliases()).To(HaveLen(0))
        err := aliasedTables.AddObjects(object)
        Expect(err).ToNot(HaveOccurred())
        Expect(aliasedTables.Aliases()).To(HaveLen(1))
    })
})
