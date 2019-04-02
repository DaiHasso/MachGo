package query

import (
    "fmt"
    "database/sql"
    "math/rand"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/jmoiron/sqlx"

    "github.com/daihasso/machgo/pool"
    "github.com/daihasso/machgo/pool/dbtype"
    qt "github.com/daihasso/machgo/query/qtypes"
    "github.com/daihasso/machgo/base"
)


type testObject struct {
    Id int64 `db:"id"`
    Name string `db:"name"`
}

type secondTestObject struct {
    Id int64 `db:"id"`
}

func (self *secondTestObject) Relationships() []base.Relationship {
    return []base.Relationship{
        base.MustRelationship(self, "id", testObject{}, "id"),
    }
}


var _ = Describe("Query", func() {
    var (
        db *sql.DB
        mock sqlmock.Sqlmock
        connPool *pool.ConnectionPool
    )
    BeforeEach(func() {
        var err error
        db, mock, err = sqlmock.New()
        Expect(err).NotTo(HaveOccurred())
        Expect(mock).ToNot(BeNil())
        dbx := sqlx.NewDb(db, "mockdb")

        connPool = &pool.ConnectionPool{
            DB: *dbx,
            Type: dbtype.Mysql,
        }
    })
    AfterEach(func() {
        db.Close()
    })

    It("should be able create", func() {
        q := NewQuery(connPool)
        Expect(q).ToNot(BeNil())
    })

    When("a Query is created", func() {
        seed := int64(52348)
        var q *Query
        BeforeEach(func() {
            rand.Seed(seed)
            q = NewQuery(connPool)
            Expect(q).ToNot(BeNil())
        })

        It("should handle a simple join", func() {
            expectedQuery := `query: '` +
                `SELECT a.id as a_id, a.name as a_name ` +
                `FROM test_objects a ` +
                `WHERE (a.id = :const_4639577150595001395)', ` +
                `args: (const_4639577150595001395: 55)`
           
            object := &testObject{
                Id: 55,
                Name: "Foo",
            }

            objColumn, err := qt.ObjectColumn(object, "id")
            Expect(err).ToNot(HaveOccurred())
            q.Join(object).Where(qt.NewDefaultCondition(
                objColumn,
                qt.InterfaceToQueryable(55),
                qt.EqualCombiner,
            ))

            queryString := q.PrintQuery()
            fmt.Fprint(GinkgoWriter, queryString)
            Expect(queryString).To(Equal(expectedQuery))
        })
        It("should be able to select a specific object", func() {
            expectedQuery := `query: 'SELECT a.id as a_id, a.name as a_name ` +
                `FROM second_test_objects b ` +
                `JOIN test_objects a ON b.id=a.id ` +
                `WHERE (a.id = :const_4639577150595001395)', ` +
                `args: (const_4639577150595001395: 55)`

            object := &testObject{}

            object2 := &secondTestObject{}

            objColumn, err := qt.ObjectColumn(object, "id")
            Expect(err).ToNot(HaveOccurred())
            q.Join(object, object2).Where(qt.NewDefaultCondition(
                objColumn,
                qt.InterfaceToQueryable(55),
                qt.EqualCombiner,
            )).Select(qt.BaseSelectable(object))

            queryString := q.PrintQuery()
            fmt.Fprint(GinkgoWriter, queryString)
            Expect(queryString).To(Equal(expectedQuery))
        })
        It("should be able to add a limit", func() {
            expectedQuery := `query: 'SELECT a.id as a_id, a.name as a_name ` +
                `FROM test_objects a ` +
                `WHERE (a.id = :const_4639577150595001395) LIMIT 10', ` +
                `args: (const_4639577150595001395: 55)`

            object := (*testObject)(nil)

            objColumn, err := qt.ObjectColumn(object, "id")
            Expect(err).ToNot(HaveOccurred())
            q.Join(object).Where(qt.NewDefaultCondition(
                objColumn,
                qt.InterfaceToQueryable(55),
                qt.EqualCombiner,
            )).Limit(10)

            queryString := q.PrintQuery()
            fmt.Fprint(GinkgoWriter, queryString)
            Expect(queryString).To(Equal(expectedQuery))
        })
        It("should be able to add an offset", func() {
            expectedQuery := `query: 'SELECT a.id as a_id, a.name as a_name ` +
                `FROM test_objects a ` +
                `WHERE (a.id = :const_4639577150595001395) OFFSET 10', ` +
                `args: (const_4639577150595001395: 55)`

            object := (*testObject)(nil)

            objColumn, err := qt.ObjectColumn(object, "id")
            Expect(err).ToNot(HaveOccurred())
            q.Join(object).Where(qt.NewDefaultCondition(
                objColumn,
                qt.InterfaceToQueryable(55),
                qt.EqualCombiner,
            )).Offset(10)

            queryString := q.PrintQuery()
            fmt.Fprint(GinkgoWriter, queryString)
            Expect(queryString).To(Equal(expectedQuery))
        })
        It("should be able to add an order by clause", func() {
            expectedQuery := `query: 'SELECT a.id as a_id, a.name as a_name ` +
                `FROM test_objects a ` +
                `WHERE (a.id = :const_4639577150595001395) ` +
                `ORDER BY a.name', ` +
                `args: (const_4639577150595001395: 55)`

            object := (*testObject)(nil)

            objColumn, err := qt.ObjectColumn(object, "id")
            Expect(err).ToNot(HaveOccurred())
            nameColumn, err := qt.ObjectColumn(object, "name")
            Expect(err).ToNot(HaveOccurred())
            q.Join(object).Where(qt.NewDefaultCondition(
                objColumn,
                qt.InterfaceToQueryable(55),
                qt.EqualCombiner,
            )).OrderBy(nameColumn)

            queryString := q.PrintQuery()
            fmt.Fprint(GinkgoWriter, queryString)
            Expect(queryString).To(Equal(expectedQuery))
        })
    })
})
