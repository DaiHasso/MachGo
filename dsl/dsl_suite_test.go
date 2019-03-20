package dsl_test

import (
    "testing"

    database "github.com/daihasso/machgo"
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    logging "github.com/daihasso/slogging"
)

var (
    expectedFormat                              = `query: "%s", args: (%+v)`
    object1, object2, object3, object4, object5 *testObject
    object6 *testObjectWithCreated
)

func TestDsl(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Dsl Package Suite")
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
        DefaultDBObject: database.DefaultDBObject{},
        table:           "testtable1",
        relationships:   make([]database.Relationship, 0),
    }

    object2 = &testObject{
        DefaultDBObject: database.DefaultDBObject{},
        table:           "testtable2",
    }

    object3 = &testObject{
        DefaultDBObject: database.DefaultDBObject{},
        table:           "testtable3",
    }

    object4 = &testObject{
        DefaultDBObject: database.DefaultDBObject{},
        table:           "testtable4",
    }

    object5 = &testObject{
        DefaultDBObject: database.DefaultDBObject{},
        table:           "testtable5",
    }

    object6 = &testObjectWithCreated{
        DefaultDBObject: database.DefaultDBObject{},
        table:           "testtable6",
    }

    object2.relationships = []database.Relationship{
        database.Relationship{
            SelfObject: object2,
            TargetObject: object1,
            SelfColumn: "foo",
            TargetColumn: "bar",
        },
    }
    object3.relationships = []database.Relationship{
        database.Relationship{
            SelfObject: object3,
            TargetObject: object1,
            SelfColumn: "baz",
            TargetColumn: "test",
        },
    }
    object4.relationships = []database.Relationship{
        database.Relationship{
            SelfObject: object4,
            TargetObject: object3,
            SelfColumn:   "baz2",
            TargetColumn: "test2",
        },
    }
    object5.relationships = []database.Relationship{
        database.Relationship{
            SelfObject: object5,
            TargetObject: object3,
            SelfColumn:   "baz3",
            TargetColumn: "test3",
        },
    }
    object6.relationships = []database.Relationship{
        database.Relationship{
            SelfObject: object6,
            TargetObject: object1,
            SelfColumn:   "baz",
            TargetColumn: "test",
        },
    }
})
