package sess_test

import (
    "database/sql"
    "errors"
    "fmt"
    "math/rand"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/jmoiron/sqlx"

    "github.com/daihasso/machgo/pool/dbtype"
    "github.com/daihasso/machgo/pool"
    . "github.com/daihasso/machgo/pool/sess"
)

var preInsertTripped = false
var idGeneratorTripped = false

type testObjectWithPreInsert struct {
    Id int64 `db:"id"`
    Name string `db:"name"`
}

func (testObjectWithPreInsert) PreInsertActions() error {
    fmt.Fprintf(GinkgoWriter, "Called PreInsertActions")
    preInsertTripped = true
    return nil
}

type testObjectWithIdGenerator struct {
    Id *int64 `db:"id"`
    Name string `db:"name"`

    testId int64
}

func (self testObjectWithIdGenerator) NewID() interface{} {
    fmt.Fprintf(GinkgoWriter, "Called NewID")
    idGeneratorTripped = true
    return &self.testId
}

var _ = Describe("SaveObject", func() {
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
        It("Should be able to save a simple object", func() {
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
        })

        It("Should be able to save an object with a custom table", func() {
            objectID := rand.Int63()
            expectedQ := `INSERT INTO test_custom_object \(id, name\) ` +
                `VALUES \(\?, \?\)`
            object := testObjectCustomTable{
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
        })

        It("Should be able to save an object with a custom ID", func() {
            objectID := rand.Int63()
            expectedQ := `INSERT INTO test_custom_object \(id, name\) ` +
                `VALUES \(\?, \?\)`
            object := testObjectCustomTable{
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
        })

        It("Should be able to save an object with a custom ID", func() {
            objectID := rand.Int63()
            expectedQ := `INSERT INTO test_custom_object \(id, name\) ` +
                `VALUES \(\?, \?\)`
            object := testObjectCustomTable{
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
        })

        It("Should call PreInsertActions when defined", func() {
            objectID := rand.Int63()
            expectedQ := `INSERT INTO test_object_with_pre_inserts ` +
                `\(id, name\) VALUES \(\?, \?\)`
            object := testObjectWithPreInsert{
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
            Expect(preInsertTripped).To(BeTrue())
            preInsertTripped = false
        })

        Context("and the id is handled by the database", func() {
            Context("and the database is mysql", func() {
                /*
                It("Should exclude id from the query and read the result " +
                    "back", func() {
                    objectID := rand.Int63()
                    expectedQ := `INSERT INTO test_object_database_ids ` +
                        `\(name\) VALUES \(\?\)`
                    object := testObjectDatabaseId{
                        Name: "foo",
                    }
                    mock.ExpectBegin()
                    mock.ExpectExec(expectedQ).WithArgs(
                        "foo",
                    ).WillReturnResult(
                        sqlmock.NewResult(objectID, 1),
                    )
                    mock.ExpectCommit()
                    err := SaveObject(&object)
                    Expect(err).ToNot(HaveOccurred())
                    Expect(object.Id).To(Equal(objectID))
                })
                */

                It("Should handle an error when reading the id from the" +
                    " result", func() {
                    expectedError := errors.New("Database explosion.")
                    expectedQ := `INSERT INTO test_object_database_ids ` +
                        `\(name\) VALUES \(\?\)`
                    object := testObjectDatabaseId{
                        Name: "foo",
                    }
                    mock.ExpectBegin()
                    mock.ExpectExec(expectedQ).WithArgs(
                        "foo",
                    ).WillReturnResult(
                        sqlmock.NewErrorResult(expectedError),
                    )
                    mock.ExpectRollback()
                    err := SaveObject(&object)
                    Expect(err).To(HaveOccurred())
                    Expect(err.Error()).To(MatchRegexp(expectedError.Error()))
                    Expect(object.Id).To(BeZero())
                })

                It("Should handle an error when executing the statement",
                    func() {
                    expectedError := errors.New("Database explosion.")
                    expectedQ := `INSERT INTO test_object_database_ids ` +
                        `\(name\) VALUES \(\?\)`
                    object := testObjectDatabaseId{
                        Name: "foo",
                    }
                    mock.ExpectBegin()
                    mock.ExpectExec(expectedQ).WithArgs(
                        "foo",
                    ).WillReturnError(expectedError)
                    mock.ExpectRollback()
                    err := SaveObject(&object)
                    Expect(err).To(HaveOccurred())
                    Expect(err.Error()).To(MatchRegexp(expectedError.Error()))
                    Expect(object.Id).To(BeZero())
                })
            })

            Context("and the database is postgres", func() {
                BeforeEach(func(){
                    dbType = dbtype.Postgres
                })
                /*
                It("Should read back the returned id", func() {
                    objectID := rand.Int63()
                    expectedQ := `INSERT INTO test_object_database_ids ` +
                        `\(name\) VALUES \(\?\) RETURNING id`
                    object := testObjectDatabaseId{
                        Name: "foo",
                    }
                    mock.ExpectBegin()
                    mock.ExpectQuery(expectedQ).WithArgs(
                        "foo",
                    ).WillReturnRows(
                        sqlmock.NewRows([]string{
                            "id",
                        }).AddRow(objectID),
                    )
                    mock.ExpectCommit()
                    err := SaveObject(&object)
                    Expect(err).ToNot(HaveOccurred())
                    Expect(object.Id).To(Equal(objectID))
                })
                */

                It("Should handle an error while reading returned id " +
                    "gracefully", func() {
                    expectedError := errors.New("Database explosion.")
                    expectedQ := `INSERT INTO test_object_database_ids ` +
                        `\(name\) VALUES \(\?\) RETURNING id`
                    object := testObjectDatabaseId{
                        Name: "foo",
                    }
                    mock.ExpectBegin()
                    mock.ExpectQuery(expectedQ).WithArgs(
                        "foo",
                    ).WillReturnError(expectedError)
                    mock.ExpectRollback()

                    err := SaveObject(&object)
                    Expect(err).To(HaveOccurred())
                    Expect(err.Error()).To(MatchRegexp(expectedError.Error()))
                    Expect(object.Id).To(Equal(int64(0)))
                })
            })

            Context("and the database type is unkown", func() {
                BeforeEach(func(){
                    dbType = dbtype.UnsetDatabaseType
                })

                It("Should handle error out gracefully", func() {
                    expectedQ := `INSERT INTO test_object_database_ids ` +
                        `\(name\) VALUES \(\?\) RETURNING id`
                    object := testObjectDatabaseId{
                        Name: "foo",
                    }
                    mock.ExpectBegin()
                    mock.ExpectQuery(expectedQ).WithArgs(
                        "foo",
                    )
                    mock.ExpectRollback()

                    err := SaveObject(&object)
                    Expect(err).To(HaveOccurred())
                    Expect(err.Error()).To(MatchRegexp("Unsupported db type"))
                    Expect(object.Id).To(Equal(int64(0)))
                })
            })
        })

        It("Should call NewID when id is not set", func() {
            objectID := rand.Int63()
            expectedQ := `INSERT INTO test_object_with_id_generators ` +
                `\(id, name\) VALUES \(\?, \?\)`
            object := testObjectWithIdGenerator{
                Id: nil,
                Name: "foo",
                testId: objectID,
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
            Expect(idGeneratorTripped).To(BeTrue())
            Expect(*object.Id).To(Equal(objectID))
            idGeneratorTripped = false
        })

        It("Should error if no ID is set and object doesn't have a generator",
            func() {
            object := testObject{
                Name: "foo",
            }
            mock.ExpectBegin()
            mock.ExpectRollback()

            err := SaveObject(&object)

            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(Equal(
                "Object has no identifier set and does not have an ID " +
                    "generator.",
            ))
        })

        It("Should allow two objects with the same id if they have " +
            "different tables", func() {
            objectID := rand.Int63()

            expectedQ := `INSERT INTO test_objects \(id, name\) ` +
                `VALUES \(\?, \?\)`
            object := testObject{
                Id: objectID,
                Name: "foo",
            }

            expectedQ2 := `INSERT INTO test_custom_object \(id, name\) ` +
                `VALUES \(\?, \?\)`
            object2 := testObjectCustomTable{
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

            mock.ExpectBegin()
            mock.ExpectExec(expectedQ2).WithArgs(
                objectID, "foo",
            ).WillReturnResult(
                    sqlmock.NewResult(objectID, 1),
            )
            mock.ExpectCommit()
            err = SaveObject(&object2)
            Expect(err).ToNot(HaveOccurred())
            Expect(Saved(&object)).To(BeTrue())
        })

        It("Should fail on an object with no discernible identifier", func() {
            object := struct{
                Name string
            }{
                Name: "test",
            }

            mock.ExpectBegin()
            mock.ExpectRollback()
            err := SaveObject(&object)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(
                MatchRegexp(
                    "Object provided to SaveObject doesn't have an " +
                    "identifier",
                ),
            )
        })

        It("Should fail on an object with an identifier but no discernible " +
            "table name", func() {
            objectID := rand.Int63()
            object := struct{
                Id int64 `db:"id"`
                Name string `db:"name"`
            }{
                Id: objectID,
                Name: "test",
            }

            mock.ExpectBegin()
            mock.ExpectRollback()
            err := SaveObject(&object)
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(
                MatchRegexp(
                    "provided has no struct name.",
                ),
            )
        })
    })
})

var _ = Describe("SaveObjects", func() {
    Context("When a global pool exists", func() {
        var (
            err error
            db *sql.DB
            mock sqlmock.Sqlmock
        )
        dbType := dbtype.Mysql
        rand.Seed(1338)
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
        It("Should be able to save multiple objects", func() {
            objectID := rand.Int63()
            object2ID := rand.Int63()
            expectedQ := `INSERT INTO test_objects \(id, name\) ` +
                `VALUES \(\?, \?\)`
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
                objectID, "foo",
            ).WillReturnResult(
                    sqlmock.NewResult(objectID, 1),
            )
            mock.ExpectCommit()
            mock.ExpectBegin()
            mock.ExpectExec(expectedQ).WithArgs(
                object2ID, "foo2",
            ).WillReturnResult(
                    sqlmock.NewResult(object2ID, 1),
            )
            mock.ExpectCommit()
            errs := SaveObjects(Objs(&object, &object2))
            Expect(errs).To(BeEmpty())
            Expect(Saved(&object)).To(BeTrue())
            Expect(Saved(&object2)).To(BeTrue())
        })

        It("Should update objects when one has already been saved", func() {
            objectID := rand.Int63()
            object2ID := rand.Int63()
            expectedQ := `INSERT INTO test_objects \(id, name\) ` +
                `VALUES \(\?, \?\)`
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
                objectID, "foo",
            ).WillReturnResult(
                    sqlmock.NewResult(objectID, 2),
            )
            mock.ExpectCommit()
            err := SaveObject(&object)
            Expect(err).ToNot(HaveOccurred())
            Expect(Saved(&object)).To(BeTrue())

            // TODO: object1 should be updated here when updateObject is
            //       complete.
            mock.ExpectBegin()
            mock.ExpectExec(expectedQ).WithArgs(
                object2ID, "foo2",
            ).WillReturnResult(
                    sqlmock.NewResult(objectID, 2),
            )
            mock.ExpectCommit()
            errs := SaveObjects(Objs(&object, &object2))
            Expect(errs).To(BeEmpty())
            Expect(Saved(&object)).To(BeTrue())
            Expect(Saved(&object2)).To(BeTrue())
        })

        It("Should not fail when not all the objects are the same type",
            func() {
            objectID := rand.Int63()
            object2ID := rand.Int63()
            expectedQ := `INSERT INTO test_objects \(id, name\) ` +
                `VALUES \(\?, \?\)`
            expectedQ2 := `INSERT INTO test_custom_object \(id, name\) ` +
                `VALUES \(\?, \?\)`
            object := testObject{
                Id: objectID,
                Name: "foo",
            }
            object2 := testObjectCustomTable{
                Id: object2ID,
                Name: "foo2",
            }

            mock.ExpectBegin()
            mock.ExpectExec(expectedQ).WithArgs(
                objectID, "foo",
            ).WillReturnResult(
                    sqlmock.NewResult(objectID, 2),
            )
            mock.ExpectCommit()
            mock.ExpectBegin()
            mock.ExpectExec(expectedQ2).WithArgs(
                object2ID, "foo2",
            ).WillReturnResult(
                    sqlmock.NewResult(objectID, 2),
            )
            mock.ExpectCommit()
            errs := SaveObjects(Objs(&object, &object2))
            Expect(errs).To(BeEmpty())
        })

        It("Should continue to save other objects when one fails", func() {
            expectedError := errors.New("Database explosion.")
            objectID := rand.Int63()
            object2ID := rand.Int63()
            object3ID := rand.Int63()
            expectedQ := `INSERT INTO test_objects \(id, name\) ` +
                `VALUES \(\?, \?\)`
            object := testObject{
                Id: objectID,
                Name: "foo",
            }
            object2 := testObject{
                Id: object2ID,
                Name: "foo2",
            }
            object3 := testObject{
                Id: object3ID,
                Name: "foo3",
            }

            mock.ExpectBegin()
            mock.ExpectExec(expectedQ).WithArgs(
                objectID, "foo",
            ).WillReturnResult(
                    sqlmock.NewResult(objectID, 2),
            )
            mock.ExpectCommit()
            mock.ExpectBegin()
            mock.ExpectExec(expectedQ).WithArgs(
                object2ID, "foo2",
            ).WillReturnError(expectedError)
            mock.ExpectRollback()
            mock.ExpectBegin()
            mock.ExpectExec(expectedQ).WithArgs(
                object3ID, "foo3",
            ).WillReturnResult(
                    sqlmock.NewResult(object3ID, 2),
            )
            mock.ExpectCommit()
            errs := SaveObjects(Objs(&object, &object2, &object3))
            Expect(errs).To(HaveLen(1))
            Expect(errs[0].Error()).To(MatchRegexp(expectedError.Error()))
            Expect(errs[0].Error()).To(MatchRegexp(
                "Error while saving object #2",
            ))
        })
    })
})
