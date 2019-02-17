package database_test

import (
	"errors"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"

    "github.com/daihasso/machgo/database/dbtype"
    "github.com/daihasso/machgo/pool"
    . "github.com/daihasso/machgo/database"
    "github.com/daihasso/machgo"
)

var _ = Describe("Manager", func() {
	setupGlobalPool := func() func() {
		return func() {
			var err error
			db, mock, err = sqlmock.New()

			Expect(err).NotTo(HaveOccurred())

			dbx = sqlx.NewDb(db, "mockdb")
			connPool := pool.ConnectionPool{
				DB: *dbx,
				Type: dbtype.Mysql,
			}

			pool.SetGlobalConnectionPool(&connPool)
			globalPool, err := pool.GlobalConnectionPool()
			Expect(globalPool).ShouldNot(BeNil())
			Expect(err).ShouldNot(HaveOccurred())
		}
	}

	Context("When no global connection pool is set", func() {
		BeforeEach(func() {
			globalPool, err := pool.GlobalConnectionPool()
			Expect(globalPool).To(BeNil())
			Expect(err).NotTo(BeNil())
		})

		It("Should fail on creation of a new manager", func() {
			manager, err := NewManager()
			Expect(err).To(HaveOccurred())
			Expect(manager).To(BeNil())
		})
	})


	Context("When a manager is created", func() {
		var manager *Manager
		BeforeEach(setupGlobalPool())
		BeforeEach(func() {
			var err error
			manager, err = NewManager()
			Expect(err).NotTo(HaveOccurred())
			Expect(manager).NotTo(BeNil())
		})
        AfterEach(func() {
            db.Close()
        })

		Context("When retrieving a single object by ID", func() {
			expectedQuery := `SELECT \* FROM fake_objects WHERE id=\?`
			expectedRows := sqlmock.NewRows([]string{"id", "name"}).AddRow(
				int64(1), "testfoobar",
			)
			It("Should handle a simple succesful case well", func() {
				mock.ExpectQuery(expectedQuery).WillReturnRows(expectedRows)

				fakeObject := FakeObject{
					Name: "foo",
				}
				id := &MachGo.IntID{1}
				err := manager.GetObject(&fakeObject, id)
				Expect(err).NotTo(HaveOccurred())
				Expect(fakeObject.GetID()).Should(Equal(id))
			})

			It("Should handle an error gracefully", func() {
				expectedErr := errors.New(
					"Oh shit dawg, this db query don blow'd up!",
				)

				mock.ExpectQuery(expectedQuery).WillReturnError(expectedErr)
				fakeObject := FakeObject{
					Name: "foo",
				}
				id := MachGo.IntID{1}
				err := manager.GetObject(&fakeObject, &id)
				Expect(err).To(Equal(expectedErr))
			})
		})

		Context("When finding an object by its attributes", func() {
			It("Should find a simple object by its name", func() {
				expectedQuery := `SELECT \* FROM fake_objects WHERE name = \?`
				expectedRows := sqlmock.NewRows([]string{"id", "name"}).AddRow(
					int64(9001), "foo",
				)
				mock.ExpectQuery(expectedQuery).WillReturnRows(expectedRows)

				fakeObject := FakeObject{
					Name: "foo",
				}
				Expect(fakeObject.IDIsSet()).To(BeFalse())
				err := manager.FindObject(&fakeObject)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeObject.GetID()).To(Equal(&MachGo.IntID{9001}))
			})

			It("Should find an object by its name and email", func() {
				expectedQuery := `SELECT \* FROM fake_complicated_objects ` +
					`WHERE \S+ = \? AND \S+ = \?`
				expectedAddress :=  "777 foobar ave."
				expectedRows := sqlmock.NewRows(
					[]string{"id", "name", "email", "address"},
				).AddRow(
					int64(9001), "foo", "foo@bar.com", expectedAddress,
				)
				mock.ExpectQuery(expectedQuery).WillReturnRows(expectedRows)

				fakeObject := FakeComplicatedObject{
					Name: "foo",
					Email: "foo@bar.com",
				}
				Expect(fakeObject.IDIsSet()).To(BeFalse())
				Expect(fakeObject.Address).To(BeEmpty())
				err := manager.FindObject(&fakeObject)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeObject.GetID()).To(Equal(&MachGo.IntID{9001}))
				Expect(fakeObject.Address).To(Equal(expectedAddress))
			})

			It("Should find an object ID and ignore other attributes if an " +
				"ID is set", func() {
				expectedQuery := `SELECT \* FROM fake_complicated_objects ` +
					`WHERE id = ?`
				expectedAddress := "777 foobar ave."
				expectedEmail := "foo@bar.com"
				expectedRows := sqlmock.NewRows(
					[]string{"id", "name", "email", "address"},
				).AddRow(
					int64(9001), "foo", expectedEmail, expectedAddress,
				)

				mock.ExpectQuery(expectedQuery).WillReturnRows(expectedRows)

				fakeObject := FakeComplicatedObject{
					Name: "foo",
				}
				err := fakeObject.SetID(&MachGo.IntID{1})
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeObject.IDIsSet()).To(BeTrue())
				Expect(fakeObject.Address).To(BeEmpty())
				Expect(fakeObject.Email).To(BeEmpty())
				err = manager.FindObject(&fakeObject)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeObject.GetID()).To(Equal(&MachGo.IntID{9001}))
				Expect(fakeObject.Address).To(Equal(expectedAddress))
				Expect(fakeObject.Email).To(Equal(expectedEmail))
			})

			It("Should handle an error gracefully", func() {
				expectedQuery := `SELECT \* FROM fake_complicated_objects ` +
					`WHERE name = ?`
				expectedErr := errors.New(
					"Oh shit dawg, this db query don blow'd up!",
				)

				mock.ExpectQuery(expectedQuery).WillReturnError(expectedErr)

				fakeObject := FakeComplicatedObject{
					Name: "foo",
				}
				err := manager.FindObject(&fakeObject)
				Expect(err).To(Equal(expectedErr))
			})
		})
		Context("When finding multiple objects", func() {
			It("Should handle an empty object as a query", func() {
				expectedQuery := `SELECT \* FROM fake_objects`
				id1 := int64(9001)
				id2 := int64(9002)
				expectedRows := sqlmock.NewRows(
					[]string{"id", "name"},
				).AddRow(
					id1, "foo",
				).AddRow(
					id2, "bar",
				)
				fakeObject := FakeObject{}

				mock.ExpectQuery(expectedQuery).WillReturnRows(expectedRows)
				results, err := manager.FindObjects(&fakeObject)
				Expect(err).ToNot(HaveOccurred())
				fakeResults, ok := results.(*[]*FakeObject)
				Expect(ok).To(BeTrue())
				Expect(len(*fakeResults)).To(Equal(2))
				Expect((*fakeResults)[0].Name).To(Equal("foo"))
				Expect((*fakeResults)[0].GetID().Value()).To(Equal(id1))
				Expect((*fakeResults)[1].Name).To(Equal("bar"))
				Expect((*fakeResults)[1].GetID().Value()).To(Equal(id2))
			})

			It("Should handle an empty composite object as a query", func() {
				expectedQuery := `SELECT \* FROM fake_composite_objects`
				expectedRows := sqlmock.NewRows(
					[]string{"name", "email", "address"},
				).AddRow(
					"foo", "foo@barmail.com", "777 foo ave.",
				).AddRow(
					"bar", "bar@foomail.com", "777 bar ave.",
				)
				fakeObject := FakeCompositeObject{}

				mock.ExpectQuery(expectedQuery).WillReturnRows(expectedRows)
				results, err := manager.FindObjects(&fakeObject)
				Expect(err).ToNot(HaveOccurred())
				fakeResults, ok := results.(*[]*FakeCompositeObject)
				Expect(ok).To(BeTrue())
				Expect(len(*fakeResults)).To(Equal(2))
				Expect((*fakeResults)[0].Name).To(Equal("foo"))
				Expect((*fakeResults)[0].Email).To(Equal("foo@barmail.com"))
				Expect((*fakeResults)[0].Address).To(Equal("777 foo ave."))
				Expect((*fakeResults)[1].Name).To(Equal("bar"))
				Expect((*fakeResults)[1].Email).To(Equal("bar@foomail.com"))
				Expect((*fakeResults)[1].Address).To(Equal("777 bar ave."))
			})

			It("Should error on an unknown object type", func() {
				_, err := manager.FindObjects(&WeirdObject{})
				Expect(err).To(HaveOccurred())
			})

			It("Should work when a composite object has both fields filled",
				func() {
				expectedQuery := `SELECT \* FROM fake_composite_objects ` +
					`WHERE [^=]+=\? AND [^=]+=\?`
				expectedRows := sqlmock.NewRows(
					[]string{"name", "email", "address"},
				).AddRow(
					"foo", "foo@barmail.com", "777 foo ave.",
				).AddRow(
					"foo", "foo@barmail.com", "777 bar ave.",
				)
				fakeObject := FakeCompositeObject{
					Name: "foo",
					Email: "foo@barmail.com",
				}
				mock.ExpectQuery(expectedQuery).WillReturnRows(expectedRows)

				results, err := manager.FindObjects(&fakeObject)
				Expect(err).ToNot(HaveOccurred())
				fakeResults, ok := results.(*[]*FakeCompositeObject)
				Expect(ok).To(BeTrue())
				Expect(len(*fakeResults)).To(Equal(2))
				Expect((*fakeResults)[0].Name).To(Equal("foo"))
				Expect((*fakeResults)[0].Email).To(Equal("foo@barmail.com"))
				Expect((*fakeResults)[0].Address).To(Equal("777 foo ave."))
				Expect((*fakeResults)[1].Name).To(Equal("foo"))
				Expect((*fakeResults)[1].Email).To(Equal("foo@barmail.com"))
				Expect((*fakeResults)[1].Address).To(Equal("777 bar ave."))
			})

			It("Should work when a composite object has only one field " +
				"filled", func() {
				expectedQuery := `SELECT \* FROM fake_composite_objects ` +
					`WHERE name=\?`
				expectedRows := sqlmock.NewRows(
					[]string{"name", "email", "address"},
				).AddRow(
					"foo", "foo@barmail.com", "777 foo ave.",
				).AddRow(
					"foo", "bar@foomail.com", "777 foo ave.",
				)
				fakeObject := FakeCompositeObject{
					Name: "foo",
				}
				mock.ExpectQuery(expectedQuery).WithArgs(
					"foo",
				).WillReturnRows(expectedRows)

				results, err := manager.FindObjects(&fakeObject)
				Expect(err).ToNot(HaveOccurred())
				fakeResults, ok := results.(*[]*FakeCompositeObject)
				Expect(ok).To(BeTrue())
				Expect(len(*fakeResults)).To(Equal(2))
				Expect((*fakeResults)[0].Name).To(Equal("foo"))
				Expect((*fakeResults)[0].Email).To(Equal("foo@barmail.com"))
				Expect((*fakeResults)[0].Address).To(Equal("777 foo ave."))
				Expect((*fakeResults)[1].Name).To(Equal("foo"))
				Expect((*fakeResults)[1].Email).To(Equal("bar@foomail.com"))
				Expect((*fakeResults)[1].Address).To(Equal("777 foo ave."))
			})

			It("Should handle an error gracefully", func() {
				expectedQuery := `SELECT \* FROM fake_objects`
				expectedErr := errors.New(
					"Oh shit dawg, this db query don blow'd up!",
				)

				mock.ExpectQuery(expectedQuery).WillReturnError(expectedErr)

				fakeObject := FakeObject{
					Name: "foo",
				}
				_, err := manager.FindObjects(&fakeObject)
				Expect(err).To(Equal(expectedErr))
			})
		})
		Context("When saving an object", func() {
			It("Should be able to save a simple object.", func() {
				expectedQuery := `INSERT INTO fake_objects \(name\) ` +
					`VALUES \(\?\)`

				fakeObject := FakeObject{
					Name: "testobj",
				}

				mock.ExpectBegin()
				mock.ExpectExec(expectedQuery).WithArgs(
					"testobj",
				).WillReturnResult(
					sqlmock.NewResult(99, 1),
				)
				mock.ExpectCommit()

				err := manager.SaveObject(&fakeObject)

				Expect(err).ToNot(HaveOccurred())
				Expect(fakeObject.GetID().Value()).To(Equal(int64(99)))
			})
		})
		Context("When updating a saved object", func() {
			It("Should be able to update a simple object.", func() {
				expectedQuery := `UPDATE fake_objects SET name=\? WHERE id=\?`

				fakeObject := FakeObject{
					Name: "testobj",
				}

				err := fakeObject.SetID(&MachGo.IntID{5})
				Expect(err).ToNot(HaveOccurred())

				fakeObject.SetSaved(true)

				mock.ExpectBegin()
				mock.ExpectExec(expectedQuery).WithArgs(
					"testobj", 5,
				).WillReturnResult(
					sqlmock.NewResult(5, 1),
				)
				mock.ExpectCommit()

				err = manager.UpdateObject(&fakeObject)

				Expect(err).ToNot(HaveOccurred())
			})

			It("Should be able to update a composite object.", func() {
				// NOTE: This really should be less permissive but ordering
				//       isn't gueranteed.
				expectedQuery := `UPDATE fake_composite_objects SET ` +
					`([^=]+=\?,? ?){3} WHERE ([^=]+=\?) AND ([^=]+=\?)`

				fakeObject := FakeCompositeObject{
					Name: "foo",
					Email: "foo@bar.com",
					Address: "777 bar court",
				}

				fakeObject.SetSaved(true)

				mock.ExpectBegin()
				// NOTE: This should really check args but same problem as
				//       above.
				mock.ExpectExec(expectedQuery).WillReturnResult(
					sqlmock.NewResult(5, 1),
				)
				mock.ExpectCommit()

				err := manager.UpdateObject(&fakeObject)

				Expect(err).ToNot(HaveOccurred())
			})

			It("Should fail on unsaved object.", func() {
				fakeObject := FakeObject{
					Name: "testobj",
				}

				mock.ExpectBegin()

				err := manager.UpdateObject(&fakeObject)

				Expect(err).To(Equal(ErrObjectNotSaved))
			})

			It("Updating an unchanged object should be a no-op.", func() {
				expectedQuery := `INSERT INTO fake_objects \(name\) ` +
					`VALUES \(\?\)`

				fakeObject := FakeObject{
					Name: "testobj",
				}

				mock.ExpectBegin()
				mock.ExpectExec(expectedQuery).WithArgs(
					"testobj",
				).WillReturnResult(
					sqlmock.NewResult(99, 1),
				)
				mock.ExpectCommit()

				err := manager.SaveObject(&fakeObject)
				Expect(err).ToNot(HaveOccurred())
				Expect(fakeObject.GetID().Value()).To(Equal(int64(99)))

				mock.ExpectBegin()
				mock.ExpectCommit()
				err = manager.UpdateObject(&fakeObject)

				Expect(err).ToNot(HaveOccurred())
			})
		})
		Context("When deleting an object", func() {
			It("Should be able to delete a simple object.", func() {
				expectedQuery := `DELETE FROM fake_objects WHERE id = \?`

				fakeObject := FakeObject{
					Name: "testobj",
				}

				err := fakeObject.SetID(&MachGo.IntID{5})
				Expect(err).ToNot(HaveOccurred())

				mock.ExpectBegin()
				mock.ExpectExec(expectedQuery).WithArgs(
					5,
				).WillReturnResult(
					sqlmock.NewResult(5, 1),
				)
				mock.ExpectCommit()

				err = manager.DeleteObject(&fakeObject)

				Expect(err).ToNot(HaveOccurred())
			})

			It("Should be able to delete a composite object.", func() {
				// NOTE: This really should be less permissive but ordering
				//       isn't gueranteed.
				expectedQuery := `DELETE FROM fake_composite_objects ` +
					`WHERE (name|email)=\? AND (name|email)=\?`

				fakeObject := FakeCompositeObject{
					Name: "foo",
					Email: "foo@bar.com",
					Address: "777 bar court",
				}

				fakeObject.SetSaved(true)

				mock.ExpectBegin()
				// NOTE: This should really check args but same problem as
				//       above.
				mock.ExpectExec(expectedQuery).WillReturnResult(
					sqlmock.NewResult(5, 1),
				)
				mock.ExpectCommit()

				err := manager.DeleteObject(&fakeObject)

				Expect(err).ToNot(HaveOccurred())
			})
		})
	})
})
