package sess_test

import (
    "database/sql"
    "math/rand"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/jmoiron/sqlx"

    "github.com/daihasso/machgo/database/dbtype"
    "github.com/daihasso/machgo/pool"
    . "github.com/daihasso/machgo/pool/sess"
)

var _ = Describe("GetObject", func() {
    Context("When a global pool exists", func() {
        var (
            err error
            db *sql.DB
            mock sqlmock.Sqlmock
        )
        dbType := dbtype.Mysql
        rand.Seed(1340)
        BeforeEach(func() {

            db, mock, err = sqlmock.New()
            Expect(err).NotTo(HaveOccurred())

        })
        JustBeforeEach(func() {
            dbx := sqlx.NewDb(db, "mockdb")
            connPool := pool.ConnectionPool{
                DB: *dbx,
                Type: dbType,
            }

            pool.SetGlobalConnectionPool(&connPool)
            globalPool, err := pool.GlobalConnectionPool()
            Expect(globalPool).ShouldNot(BeNil())
            Expect(err).Should(BeNil())
        })
        AfterEach(func() {
            db.Close()
        })

        It("Should be able to get a simple object", func() {
            objectID := rand.Int63()
            expectedQ := `SELECT \* FROM test_objects WHERE id = @id`
            object := testObject{}
            mock.ExpectBegin()
            mock.ExpectQuery(expectedQ).WithArgs(
                objectID,
            ).WillReturnRows(
                sqlmock.NewRows(
                    []string{"id", "name"},
                ).AddRow(
                    int64(objectID), "foo",
                ),
            )
            mock.ExpectCommit()
            err := GetObject(&object, objectID)
            Expect(err).ToNot(HaveOccurred())
            Expect(object.Name).To(Equal("foo"))
            Expect(object.Id).To(Equal(objectID))
        })

        It("Should fail when Id is set", func() {
            expectedError := "Object provided to GetObject has an " +
                "identifier set, it should be a new instance with no" +
                " identifier."
            object := testObject{
                Id: 5,
            }
            mock.ExpectBegin()
            mock.ExpectRollback()
            err := GetObject(&object, 5)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(Equal(expectedError))
        })

        It("Should fail when ID doesn't exist", func() {
            expectedError := "Object provided to GetObject doesn't have " +
                "an identifier."
            object := struct{
                Name string
            }{
                Name: "test",
            }
            mock.ExpectBegin()
            mock.ExpectRollback()
            err := GetObject(&object, 1)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(Equal(expectedError))
            Expect(Saved(&object)).To(BeFalse())
        })
    })
})
