package qtypes

import (
    "database/sql"
    "fmt"
    "math/rand"
    "reflect"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/jmoiron/sqlx"

    "github.com/daihasso/machgo/refl"
)

type testObjectQR struct {
    Id int64 `db:"id"`
    Name string
}

var _ = Describe("QueryResults", func() {
    var (
        db *sql.DB
        dbx *sqlx.DB
        mock sqlmock.Sqlmock
    )
    rand.Seed(84842)
    BeforeEach(func() {
        var err error
        db, mock, err = sqlmock.New()
        Expect(err).NotTo(HaveOccurred())
    })
    JustBeforeEach(func() {
        dbx = sqlx.NewDb(db, "mockdb")
    })
    AfterEach(func() {
        db.Close()
    })

    It("should be able to be created", func() {
        expectedRows := sqlmock.NewRows(
            []string{"id", "name"},
        ).AddRow(1, "foo")
        mock.ExpectBegin()
        mock.ExpectQuery("SELECT").WillReturnRows(expectedRows)
        tx, err := dbx.Beginx()
        Expect(err).ToNot(HaveOccurred())

        rows, err := tx.Queryx("SELECT")
        Expect(err).ToNot(HaveOccurred())

        object := testObjectQR{}

        at, err := NewAliasedTables(&object)
        Expect(err).ToNot(HaveOccurred())
        typeBSFieldMap := make(map[reflect.Type]*refl.GroupedFieldsWithBS)
        objType := refl.Deref(reflect.TypeOf(object))
        fieldGroupings := refl.GetGroupedFieldsWithBS(
            object,
            refl.GroupFieldsByTagValue("db", "dbfkey"),
        )
        tagValBSFields := fieldGroupings[0]
        typeBSFieldMap[objType] = tagValBSFields

        qr := NewQueryResults(tx, rows, at, typeBSFieldMap)
        Expect(qr).ToNot(BeNil())
    })
    It("should be able to write all results", func() {
        expectedId, expectedName := int64(1), "foo"
        expectedRows := sqlmock.NewRows(
            []string{"a_id", "a_name"},
        ).AddRow(expectedId, expectedName)
        mock.ExpectBegin()
        mock.ExpectQuery("SELECT").WillReturnRows(expectedRows)
        mock.ExpectCommit()
        tx, err := dbx.Beginx()
        Expect(err).ToNot(HaveOccurred())

        rows, err := tx.Queryx("SELECT")
        Expect(err).ToNot(HaveOccurred())

        object := &testObjectQR{}

        at, err := NewAliasedTables(object)
        Expect(err).ToNot(HaveOccurred())

        typeBSFieldMap := make(map[reflect.Type]*refl.GroupedFieldsWithBS)
        objType := refl.Deref(reflect.TypeOf(object))
        fieldGroupings := refl.GetGroupedFieldsWithBS(
            object,
            refl.GroupFieldsByTagValue("db", "dbfkey"),
        )
        tagValBSFields := fieldGroupings[0]
        typeBSFieldMap[objType] = tagValBSFields

        qr := NewQueryResults(tx, rows, at, typeBSFieldMap)
        Expect(qr).ToNot(BeNil())

        results := make([]*testObjectQR, 0)
        err = qr.WriteAllTo(&results)
        Expect(err).ToNot(HaveOccurred())

        fmt.Fprintf(GinkgoWriter, "%#+v\n", results)

        Expect(results[0].Id).To(Equal(expectedId))
        Expect(results[0].Name).To(Equal(expectedName))
    })
})
