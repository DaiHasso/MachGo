package  sess_test

import (
    "database/sql"
    "math/rand"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/jmoiron/sqlx"

    "github.com/daihasso/machgo/pool/dbtype"
    "github.com/daihasso/machgo/pool"
    . "github.com/daihasso/machgo/pool/sess"
)

var _ = Describe("UpdateObject", func() {
    Context("When a global pool exists", func() {
        var (
            err error
            db *sql.DB
            mock sqlmock.Sqlmock
        )
        dbType := dbtype.Mysql
        rand.Seed(1339)
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

        It("Should be able to update a simple object", func() {
            objectID := rand.Int63()
            expectedQ := `UPDATE test_objects SET id = \?, name = \? ` +
                `WHERE \(id = \?\)`
            object := testObject{
                Id: objectID,
                Name: "foo",
            }
            mock.ExpectBegin()
            mock.ExpectExec(expectedQ).WithArgs(
                objectID, "foo", objectID,
            ).WillReturnResult(
                    sqlmock.NewResult(objectID, 1),
            )
            mock.ExpectCommit()
            err := UpdateObject(&object)
            Expect(err).ToNot(HaveOccurred())
            Expect(Saved(&object)).To(BeTrue())
        })

        It("Should fail when Id is unset", func() {
            expectedError := "Object provided to UpdateObject has an " +
                "identifier but it hasn't been set."
            object := testObject{
                Name: "foo",
            }
            mock.ExpectBegin()
            mock.ExpectRollback()
            err := UpdateObject(&object)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(Equal(expectedError))
            Expect(Saved(&object)).To(BeFalse())
        })

        It("Should fail when ID doesn't exist", func() {
            expectedError := "Object provided to UpdateObject doesn't have " +
                "an identifier."
            object := struct{
                Name string
            }{
                Name: "test",
            }
            mock.ExpectBegin()
            mock.ExpectRollback()
            err := UpdateObject(&object)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(Equal(expectedError))
            Expect(Saved(&object)).To(BeFalse())
        })
    })
})
