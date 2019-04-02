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

var _ = Describe("Global Session State", func() {
    Context("When a global pool exists", func() {
        var (
            err error
            db *sql.DB
            mock sqlmock.Sqlmock
        )
        dbType := dbtype.Mysql
        rand.Seed(1337)
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

        It("Should detect a changed object", func() {
            objectID := rand.Int63()
            expectedQ := `INSERT INTO test_objects \(id, name\) ` +
                `VALUES \(\?, \?\)`
            object := testObject{
                Id: objectID,
                Name: "foo",
            }
            mock.ExpectBegin()
            mock.ExpectExec(expectedQ).WithArgs(
                objectID, "foo",
            ).WillReturnResult(
                    sqlmock.NewResult(objectID, 1),
            )
            mock.ExpectCommit()
            err := SaveObject(&object)
            Expect(err).ToNot(HaveOccurred())
            Expect(Saved(&object)).To(BeTrue())

            Expect(ObjectChanged(&object)).To(BeFalse())

            object.Name = "foobar"

            Expect(ObjectChanged(&object)).To(BeTrue())
        })

        It("Should mark an unsaved object as changed", func() {
            objectID := rand.Int63()
            object := testObject{
                Id: objectID,
                Name: "foo",
            }
            Expect(ObjectChanged(&object)).To(BeTrue())
        })

        It("Should error when checking if an object without an identifier " +
            "has changed", func() {

            object := struct{
                Name string
            }{
                Name: "test",
            }
            checkObjectChanged := func () {
                ObjectChanged(&object)
            }
            Expect(checkObjectChanged).To(Panic())
        })
    })
})
