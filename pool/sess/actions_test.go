package sess_test

import (
    "database/sql"
	"fmt"
	"math/rand"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"

    "MachGo/database/dbtype"
    "MachGo/pool"
    . "MachGo/pool/sess"
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

var _ = Describe("Session Actions", func() {
	Context("When a global pool exists", func() {
        var (
            err error
            db *sql.DB
            mock sqlmock.Sqlmock
        )
        BeforeEach(func() {
            rand.Seed(1337)

            db, mock, err = sqlmock.New()
            Expect(err).NotTo(HaveOccurred())
            dbx := sqlx.NewDb(db, "mockdb")

            connPool := pool.ConnectionPool{
                DB: *dbx,
                Type: dbtype.Mysql,
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
				`VALUES \(@id, @name\)`
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
			Expect(Saved(object)).To(BeTrue())
		})

		It("Should be able to save an object with a custom table", func() {
			objectID := rand.Int63()
			expectedQ := `INSERT INTO test_custom_object \(id, name\) ` +
				`VALUES \(@id, @name\)`
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
			expectedQ := `INSERT INTO test_object_custom_ids \(id, name\) ` +
				`VALUES \(@id, @name\)`
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
			expectedQ := `INSERT INTO test_object_custom_ids \(id, name\) ` +
				`VALUES \(@id, @name\)`
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
				`\(id, name\) VALUES \(@id, @name\)`
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

		It("Should exclude id when it's handled by the database", func() {
			objectID := rand.Int63()
			expectedQ := `INSERT INTO test_object_database_ids ` +
				`\(name\) VALUES \(@name\)`
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
			Expect(*object.Id).To(Equal(objectID))
		})

		It("Should call NewID when id is not set", func() {
			objectID := rand.Int63()
			expectedQ := `INSERT INTO test_object_with_id_generators ` +
				`\(id, name\) VALUES \(@id, @name\)`
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

		It("Should allow two objects with the same id if they have " +
			"different tables", func() {
			objectID := rand.Int63()

			expectedQ := `INSERT INTO test_objects \(id, name\) ` +
				`VALUES \(@id, @name\)`
			object := testObject{
				Id: objectID,
				Name: "foo",
			}

			expectedQ2 := `INSERT INTO test_custom_object \(id, name\) ` +
				`VALUES \(@id, @name\)`
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
			Expect(Saved(object)).To(BeTrue())

			mock.ExpectBegin()
			mock.ExpectExec(expectedQ2).WithArgs(
				objectID, "foo",
			).WillReturnResult(
					sqlmock.NewResult(objectID, 1),
			)
			mock.ExpectCommit()
			err = SaveObject(&object2)
			Expect(err).ToNot(HaveOccurred())
			Expect(Saved(object)).To(BeTrue())
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
				Id int64
				Name string
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
