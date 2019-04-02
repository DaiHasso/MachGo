package actrec

import (
    "database/sql"
    "math/rand"
    "fmt"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/jmoiron/sqlx"

    "github.com/daihasso/machgo/pool/dbtype"
    "github.com/daihasso/machgo/pool"
    "github.com/daihasso/machgo/pool/sess"
)

type ARTestObject struct {
    Record

    Id int64 `db:"id"`
    Name string `db:"name"`
}

var TheARTestObject = ARTestObject{}

var _ = Describe("ActiveRecordLinker", func() {
    BeforeEach(func() {
        err := LinkActiveRecord(&TheARTestObject)
        Expect(err).ToNot(HaveOccurred())
    })
   
    It("should be able to link an active record to an instance.", func() {
        object := ARTestObject{}
        Expect(object.Record).To(BeNil())
        err := LinkActiveRecord(&object)
        Expect(err).ToNot(HaveOccurred())
        Expect(object.Record).ToNot(BeNil())
    })

    It("should error when an object is passed by value", func() {
        object := ARTestObject{}
        err := LinkActiveRecord(object)
        Expect(err).To(HaveOccurred())
    })

    It("should error when an object is a pointer to a pointer", func() {
        object := ARTestObject{}
        objPtr := &object
        err := LinkActiveRecord(&objPtr)
        Expect(err).To(HaveOccurred())
    })

    When("an object has been linked", func() {
        var (
            db *sql.DB
            mock sqlmock.Sqlmock
            object ARTestObject
            dbType = dbtype.Mysql
        )

        BeforeEach(func() {
            var err error
            db, mock, err = sqlmock.New()
            Expect(err).NotTo(HaveOccurred())

            object = ARTestObject{}

            err = LinkActiveRecord(&object)
            Expect(err).ToNot(HaveOccurred())
            fmt.Fprint(
                GinkgoWriter,
                object.Record,
            )
        })
        JustBeforeEach(func() {
            rand.Seed(1337)
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

        It("should be able to save", func() {
            objectID := rand.Int63()
            expectedQ := `INSERT INTO ar_test_objects \(id, name\) ` +
                `VALUES \(\?, \?\)`

            object.Id = objectID
            object.Name = "foo"
            mock.ExpectBegin()
            mock.ExpectExec(expectedQ).WithArgs(
                objectID, "foo",
            ).WillReturnResult(
                    sqlmock.NewResult(objectID, 1),
            )
            mock.ExpectCommit()
            err := object.Save()
            Expect(err).ToNot(HaveOccurred())
            Expect(sess.Saved(&object)).To(BeTrue())
        })

        It("should be able to save", func() {
            objectID := rand.Int63()
            expectedQ := `INSERT INTO ar_test_objects \(id, name\) ` +
                `VALUES \(\?, \?\)`

            object.Id = objectID
            object.Name = "foo"
            mock.ExpectBegin()
            mock.ExpectExec(expectedQ).WithArgs(
                objectID, "foo",
            ).WillReturnResult(
                    sqlmock.NewResult(objectID, 1),
            )
            mock.ExpectCommit()
            err := object.Save()
            Expect(err).ToNot(HaveOccurred())
            Expect(sess.Saved(&object)).To(BeTrue())
        })
    })
})
