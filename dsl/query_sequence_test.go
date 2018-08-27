package dsl_test

import (
    "database/sql"
    "database/sql/driver"
    "fmt"
    "time"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    sqlmock "github.com/DATA-DOG/go-sqlmock"
    "github.com/DaiHasso/MachGo/database"
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

        It("Should be able to return all data as interfaces for joined objects",
            func() {
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
            },
        )
        It("Should be able to to return data into an object", func() {
            qs := dsl.NewJoin(
                object1,
                object2,
            ).SelectObject(
                object1,
            ).SetManager(manager)
            expectedQ := `SELECT a.* FROM testtable2 b JOIN ` +
                `testtable1 a ON b\.foo=a\.bar`

            expectedRow1 := []driver.Value{1, "foo"}
            expectedRow2 := []driver.Value{2, "bar"}
            allIDs := []driver.Value{expectedRow1[0], expectedRow2[0]}
            expectedRows := sqlmock.NewRows([]string{
                "a.id",
                "a.name",
            }).AddRow(expectedRow1...).AddRow(expectedRow2...)
            mock.ExpectBegin()
            mock.ExpectQuery(expectedQ).WillReturnRows(expectedRows)
            mock.ExpectCommit()

            values, err := qs.IntoObjects()
            Expect(err).NotTo(HaveOccurred())

            for i, result := range(values) {
                Expect(len(result)).To(
                    Equal(1),
                    "Should only have one result per row.",
                )
                testObj := (result[0]).(*testObject)
                fmt.Fprintf(GinkgoWriter, "Result %d:\n%+v\n", i, testObj)
                Expect(allIDs).To(ContainElement(testObj.Id))
            }
        })
        It("Should error on erroneous data", func() {
            qs := dsl.NewJoin(
                object1,
                object2,
            ).SelectObject(
                object1,
            ).SetManager(manager)
            expectedQ := `SELECT a.* FROM testtable2 b JOIN ` +
                `testtable1 a ON b\.foo=a\.bar`

            expectedRow1 := []driver.Value{1, 666}
            expectedRow2 := []driver.Value{2, 666}
            expectedRows := sqlmock.NewRows([]string{
                "a.id",
                "c.foo",
            }).AddRow(expectedRow1...).AddRow(expectedRow2...)
            mock.ExpectBegin()
            mock.ExpectQuery(expectedQ).WillReturnRows(expectedRows)
            mock.ExpectRollback()

            values, err := qs.IntoObjects()
            Expect(err).To(HaveOccurred())
            Expect(err.Error()).To(
                MatchRegexp(
                    `An object with the alias '[^']+' hasn't been added to ` +
                        `the query.`,
                ),
            )

            for i, result := range(values) {
                Expect(len(result)).To(Equal(1))
                testObj := (result[0]).(*testObject)
                Expect(testObj.Id).To(Equal(i+1))
                fmt.Fprintf(GinkgoWriter, "Result %d:\n%+v\n", i, testObj)
            }
        })
        It("Multiple ojects should be sorted in select order", func() {
            qs := dsl.NewJoin(
                object1,
                object2,
                object6,
            ).SelectObject(
                object1,
                object6,
                object2,
            ).SetManager(manager)
            expectedQ := `SELECT a.* FROM testtable2 b JOIN ` +
                `testtable1 a ON b\.foo=a\.bar`

            expectedRow1 := []driver.Value{
                1,
                "bar",
                time.Date(2001, time.November, 13, 0, 0, 0, 0, time.UTC),
            }
            expectedRow2 := []driver.Value{
                2,
                "foo",
                time.Date(2004, time.November, 17, 0, 0, 0, 0, time.UTC),
            }
            expectedRows := sqlmock.NewRows([]string{
                "a.id",
                "b.name",
                "c.created",
            }).AddRow(expectedRow1...).AddRow(expectedRow2...)
            mock.ExpectBegin()
            mock.ExpectQuery(expectedQ).WillReturnRows(expectedRows)
            mock.ExpectCommit()

            values, err := qs.IntoObjects()
            Expect(err).NotTo(HaveOccurred())

            for i, result := range(values) {
                Expect(len(result)).To(Equal(3))
                fmt.Fprintf(GinkgoWriter, "Result %d:\n%+v\n", i, result)
                for i, obj := range result {
                    if i == 1 {
                        testObj, ok := obj.(*testObjectWithCreated)
                        Expect(ok).To(
                            Equal(true),
                            "Second object isn't a testObjectWithCreated.",
                        )
                        fmt.Fprintf(GinkgoWriter, "%+v\n", testObj)
                    } else {
                        testObj, ok := obj.(*testObject)
                        Expect(ok).To(
                            Equal(true),
                            "Other object isn't a testObject.",
                        )
                        fmt.Fprintf(GinkgoWriter, "%+v\n", testObj)
                    }
                }
            }
        })
    })
})
