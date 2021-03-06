package actrec

import (
	"testing"

    "github.com/daihasso/slogging"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestActrec(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Actrec Suite")
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
