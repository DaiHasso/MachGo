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
        expectedQuery := "SELECT a.name as a_name, a.id as a_id, "+
            "a.created as a_created, a.updated a_updated, "+
            "c.name as c_name, c.id as c_id, c.created as c_created, "+
            "c.updated c_updated "+
            "FROM testtable2 b "+
            "JOIN testtable1 a ON b.foo=a.bar "+
            "JOIN testtable3 c ON a.test=c.baz"
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
        expectedQuery := "SELECT a.foo, b.bar "+
            "FROM testtable2 b "+
            "JOIN testtable1 a ON b.foo=a.bar "+
            "JOIN testtable3 c ON a.test=c.baz "+
            "JOIN testtable4 d ON c.test2=d.baz2 "+
            "JOIN testtable5 e ON c.test3=e.baz3"
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

        It("Should be able to return data as interfaces for joined objects",
            func() {
                qs := dsl.NewJoin(
                    object1,
                    object2,
                    object3,
                    object4,
                    object5,
                ).SetManager(manager)

                expectedQ := `SELECT a\.\*, b\.\*, c\.\*, d\.\*, e\.\* ` +
                    "FROM testtable2 b JOIN testtable1 a ON b.foo=a.bar "+
                    "JOIN testtable3 c ON a.test=c.baz "+
                    "JOIN testtable4 d ON c.test2=d.baz2 "+
                    "JOIN testtable5 e ON c.test3=e.baz3"

                expectedRow1 := []driver.Value{
                    int64(1), int64(1), int64(1), int64(1), int64(1),
                }
                expectedRow2 := []driver.Value{
                    int64(2), int64(2), int64(2), int64(2), int64(2),
                }
                allExpected := [][]driver.Value{expectedRow1, expectedRow2}
                expectedRows := sqlmock.NewRows([]string{
                    "b_id",
                    "a_id",
                    "c_id",
                    "d_id",
                    "e_id",
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
            expectedQ := "SELECT a.name as a_name, a.id as a_id, "+
                "a.created as a_created, a.updated a_updated "+
                "FROM testtable2 b "+
                "JOIN testtable1 a ON b.foo=a.bar"

            expectedRow1 := []driver.Value{int64(1), "foo"}
            expectedRow2 := []driver.Value{int64(2), "bar"}
            allIDs := []driver.Value{expectedRow1[0], expectedRow2[0]}
            expectedRows := sqlmock.NewRows([]string{
                "a_id",
                "a_name",
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
                idVal, err := testObj.GetID().Value()
                Expect(err).NotTo(HaveOccurred())
                Expect(allIDs).To(ContainElement(idVal))
            }
        })
        It("Should error on erroneous data", func() {
            qs := dsl.NewJoin(
                object1,
                object2,
            ).SelectObject(
                object1,
            ).SetManager(manager)
            expectedQ := "SELECT a.name as a_name, a.id as a_id, "+
                "a.created as a_created, a.updated a_updated "+
                "FROM testtable2 b JOIN testtable1 a ON b.foo=a.bar"

            expectedRow1 := []driver.Value{int64(1), int64(666)}
            expectedRow2 := []driver.Value{int64(2), int64(666)}
            expectedRows := sqlmock.NewRows([]string{
                "a_id",
                "c_foo",
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
                Expect(testObj.GetID()).To(Equal(i+1))
                fmt.Fprintf(GinkgoWriter, "Result %d:\n%+v\n", i, testObj)
            }
        })
        It("Multiple ojects should be sorted in select order", func() {
            qs := dsl.NewJoin(
                object1,
                object2,
                object6,
                object3,
            ).SelectObject(
                object1,
                object6,
                object2,
            ).SetManager(manager)
            expectedQ := "SELECT a.*, c.*, b.* FROM testtable2 b " +
                "JOIN testtable1 a ON b.foo=a.bar " +
                "JOIN testtable6 c ON a.test=c.baz " +
                "JOIN testtable3 d ON a.test=d.baz"

            expectedRow1 := []driver.Value{
                int64(1),
                "bar",
                time.Date(2001, time.November, 13, 0, 0, 0, 0, time.UTC),
            }
            expectedRow2 := []driver.Value{
                int64(2),
                "foo",
                time.Date(2004, time.November, 17, 0, 0, 0, 0, time.UTC),
            }
            expectedRows := sqlmock.NewRows([]string{
                "a_id",
                "b_name",
                "c_created",
            }).AddRow(expectedRow1...).AddRow(expectedRow2...)
            mock.ExpectBegin()
            mock.ExpectQuery(expectedQ).WillReturnRows(expectedRows)
            mock.ExpectCommit()

            outputString := qs.PrintQuery()
            fmt.Fprint(GinkgoWriter, outputString)

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
