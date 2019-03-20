package sess_test

import (
    "testing"

    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    logging "github.com/daihasso/slogging"

    "github.com/daihasso/machgo/base"
)

type testObject struct {
    Id int64 `db:"id"`
    Name string `db:"name"`
}

type testObjectCustomTable struct {
    Id int64 `db:"id"`
    Name string `db:"name"`
}

func (testObjectCustomTable) TableName() string {
    return "test_custom_object"
}

type testObjectDatabaseId struct {
    base.DatabaseManagedID

    Id int64 `db:"id"`
    Name string `db:"name"`
}


func TestDsl(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Sess Package Suite")
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
})
