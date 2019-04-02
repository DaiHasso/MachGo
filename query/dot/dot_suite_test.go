package dot_test

import (
    "testing"

    logging "github.com/daihasso/slogging"
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

    "github.com/daihasso/machgo"
)

var object1, object2, object3, object4, object5 *testObject

func TestDot(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Dot Package Suite")
}

type testObject struct {
    machgo.DefaultDBObject
    table         string
    relationships []machgo.Relationship
    Name string `db:"name"`
}

func (self *testObject) GetTableName() string {
    return self.table
}

func (self *testObject) Relationships() []machgo.Relationship {
    return self.relationships
}

var _ = BeforeSuite(func() {
    logger, err := logging.NewLogger(
        "tests",
        logging.WithLogWriters(GinkgoWriter),
        logging.WithLogLevel(logging.DEBUG),
        logging.WithFormat(logging.Standard),
    )
    if err != nil {
        panic(err)
    }

    err = logging.SetRootLogger("tests", logger)
    if err != nil {
        panic(err)
    }

    object1 = &testObject{
        DefaultDBObject: machgo.DefaultDBObject{},
        table:           "testtable1",
        relationships:   make([]machgo.Relationship, 0),
    }

    object2 = &testObject{
        DefaultDBObject: machgo.DefaultDBObject{},
        table:           "testtable2",
    }

    object3 = &testObject{
        DefaultDBObject: machgo.DefaultDBObject{},
        table:           "testtable3",
    }

    object4 = &testObject{
        DefaultDBObject: machgo.DefaultDBObject{},
        table:           "testtable4",
    }

    object5 = &testObject{
        DefaultDBObject: machgo.DefaultDBObject{},
        table:           "testtable5",
    }

    object2.relationships = []machgo.Relationship{
        machgo.Relationship{
            SelfObject: object2,
            TargetObject: object1,
            SelfColumn: "foo",
            TargetColumn: "bar",
        },
    }
    object3.relationships = []machgo.Relationship{
        machgo.Relationship{
            SelfObject: object3,
            TargetObject: object1,
            SelfColumn: "baz",
            TargetColumn: "test",
        },
    }
    object4.relationships = []machgo.Relationship{
        machgo.Relationship{
            SelfObject: object4,
            TargetObject: object3,
            SelfColumn:   "baz2",
            TargetColumn: "test2",
        },
    }
    object5.relationships = []machgo.Relationship{
        machgo.Relationship{
            SelfObject: object5,
            TargetObject: object3,
            SelfColumn:   "baz3",
            TargetColumn: "test3",
        },
    }
})
