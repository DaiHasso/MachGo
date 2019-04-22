package sess_test

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

var _ = Describe("DeleteObject", func() {
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

        It("Should be able to delete a simple object", func() {
            objectID := rand.Int63()
            expectedQ := `DELETE FROM test_objects WHERE id = ?`
            object := testObject{
                Id: objectID,
                Name: "foo",
            }
            mock.ExpectBegin()
            mock.ExpectExec(expectedQ).WithArgs(
                objectID,
            ).WillReturnResult(
                    sqlmock.NewResult(objectID, 1),
            )
            mock.ExpectCommit()
            err := DeleteObject(&object)
            Expect(err).ToNot(HaveOccurred())
        })

        It("Should fail when Id is unset", func() {
            expectedError := "Object provided to DeleteObject has an " +
                "identifier but it hasn't been set."
            object := testObject{
                Name: "foo",
            }
            mock.ExpectBegin()
            mock.ExpectRollback()
            err := DeleteObject(&object)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(Equal(expectedError))
            Expect(Saved(&object)).To(BeFalse())
        })

        It("Should fail when ID doesn't exist", func() {
            expectedError := "Object provided to DeleteObject doesn't have " +
                "an identifier."
            object := struct{
                Name string
            }{
                Name: "test",
            }
            mock.ExpectBegin()
            mock.ExpectRollback()
            err := DeleteObject(&object)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(Equal(expectedError))
            Expect(Saved(&object)).To(BeFalse())
        })

        It("Should be able to delete multiple objects", func() {
            objectID := rand.Int63()
            object2ID := rand.Int63()
            expectedQ := `DELETE FROM test_objects WHERE id = ?`
            object := testObject{
                Id: objectID,
                Name: "foo",
            }
            object2 := testObject{
                Id: object2ID,
                Name: "foo2",
            }
            mock.ExpectBegin()
            mock.ExpectExec(expectedQ).WithArgs(
                objectID,
            ).WillReturnResult(
                    sqlmock.NewResult(objectID, 1),
            )
            mock.ExpectCommit()
            mock.ExpectBegin()
            mock.ExpectExec(expectedQ).WithArgs(
                object2ID,
            ).WillReturnResult(
                    sqlmock.NewResult(object2ID, 1),
            )
            mock.ExpectCommit()
            errs := DeleteObjects(Objs(&object, &object2))
            Expect(errs).To(BeEmpty())
        })
    })
})
