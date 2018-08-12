package dsl_test

import (
	"database/sql"
	"database/sql/driver"
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	database "github.com/DaiHasso/MachGo"
	"github.com/DaiHasso/MachGo/dsl"
)

var _ = Describe("QuerySequence", func() {
	It("Should generate a query string from a join", func() {
		expectedQuery := "SELECT a.*, b.* FROM testtable2 b JOIN " +
			"testtable1 a ON b.foo=a.bar"
		expected := fmt.Sprintf(
			expectedFormat,
			expectedQuery,
			[]string{},
		)
		qs := dsl.NewJoin(object1).Join(object2)
		outputString := qs.PrintQuery()
		fmt.Fprint(GinkgoWriter, outputString)
		Expect(outputString).To(Equal(expected))
	})

	It("Should be able to override objects in the select", func() {
		expectedQuery := "SELECT a.*, c.* FROM testtable2 b JOIN " +
			"testtable1 a ON b.foo=a.bar JOIN testtable3 c ON " +
			"a.baz=c.test"
		expected := fmt.Sprintf(
			expectedFormat,
			expectedQuery,
			[]string{},
		)
		qs := dsl.NewJoin(
			object1,
			object2,
			object3,
		).SelectObject(
			object1,
			object3,
		)
		outputString := qs.PrintQuery()
		fmt.Fprint(GinkgoWriter, outputString)
		Expect(outputString).To(Equal(expected))
	})
	It("Should be able to override the select explicitly", func() {
		expectedQuery := "SELECT a.foo, b.bar FROM testtable2 b " +
			"JOIN testtable1 a ON b.foo=a.bar JOIN testtable3 c ON " +
			"a.baz=c.test JOIN testtable4 d ON c.baz2=d.test2 JOIN " +
			"testtable5 e ON c.baz3=e.test3"
		expected := fmt.Sprintf(
			expectedFormat,
			expectedQuery,
			[]string{},
		)
		qs := dsl.NewJoin(object1).Join(
			object2,
			object3,
			object4,
			object5,
		).Select(
			"testtable1.foo",
			"testtable2.bar",
		)
		outputString := qs.PrintQuery()
		fmt.Fprint(GinkgoWriter, outputString)
		Expect(outputString).To(Equal(expected))
	})
	It("Should be able to override the select explicitly", func() {
		expectedQuery := "SELECT a.foo, b.bar FROM testtable2 b " +
			"JOIN testtable1 a ON b.foo=a.bar JOIN testtable3 c ON " +
			"a.baz=c.test JOIN testtable4 d ON c.baz2=d.test2 JOIN " +
			"testtable5 e ON c.baz3=e.test3"
		expected := fmt.Sprintf(
			expectedFormat,
			expectedQuery,
			[]string{},
		)
		qs := dsl.NewJoin(object1).Join(
			object2,
			object3,
			object4,
			object5,
		).Select(
			"testtable1.foo",
			"testtable2.bar",
		)
		outputString := qs.PrintQuery()
		fmt.Fprint(GinkgoWriter, outputString)
		Expect(outputString).To(Equal(expected))
	})
	Context("When a manager is set", func() {
		var (
			err     error
			manager *database.Manager
			db      *sql.DB
			mock    sqlmock.Sqlmock
		)
		BeforeEach(func() {
			db, mock, err = sqlmock.New()
			Expect(err).NotTo(HaveOccurred())

			manager, err = database.NewManagerFromExisting(
				database.Mysql,
				db,
				"mockdb",
			)
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			db.Close()
		})

		It("Should return all data for joined objects", func() {
			qs := dsl.NewJoin(
				object1,
				object2,
				object3,
				object4,
				object5,
			).SetManager(manager)

			expectedQ := `SELECT a.*, b.*, c.*, d.*, e.* FROM testtable2 b JOIN ` +
				`testtable1 a ON b\.foo=a\.bar JOIN testtable3 c ON a\.baz=c\.test ` +
				`JOIN testtable4 d ON c\.baz2=d\.test2 JOIN testtable5 e ON ` +
				`c.baz3=e\.test3`

			expectedRow1 := []driver.Value{1, 1, 1, 1, 1}
			expectedRow2 := []driver.Value{2, 2, 2, 2, 2}
			allExpected := [][]driver.Value{expectedRow1, expectedRow2}
			expectedRows := sqlmock.NewRows([]string{
				"b.id",
				"a.id",
				"c.id",
				"d.id",
				"e.id",
			}).AddRow(expectedRow1...).AddRow(expectedRow2...)
			mock.ExpectBegin()
			mock.ExpectQuery(expectedQ).WillReturnRows(expectedRows)
			mock.ExpectCommit()

			results, err := qs.QueryInterface()
			Expect(err).NotTo(HaveOccurred())

			fmt.Fprint(GinkgoWriter, results)

			for i, result := range results {
				for j, col := range result {
					expected := allExpected[i][j]
					Expect(col).To(Equal(expected))
				}
			}
		})
	})
})
