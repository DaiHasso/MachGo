package query

import (
    "testing"
   
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"
    "github.com/daihasso/slogging"
)

func TestDsl(t *testing.T) {
    RegisterFailHandler(Fail)
    RunSpecs(t, "Query Package Suite")
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
