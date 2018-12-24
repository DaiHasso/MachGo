package sess_test

import (
	"testing"
	"log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/daihasso/slogging"

	"MachGo/base"
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

type testObjectCustomId struct {
	CustomId int64
}

func (self testObjectCustomId) SetID(in interface{}) error {
	self.CustomId = in.(int64)
	return nil
}

func (self testObjectCustomId) ID() (interface{}, bool) {
	set := true
	if self.CustomId == 0 {
		set = false
	}
	return self.CustomId, set
}

type testObjectDatabaseId struct {
	base.DatabaseManagedID

	Id *int64 `db:"id"`
	Name string `db:"name"`
}


func TestDsl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dsl Package Suite")
}

var _ = BeforeSuite(func() {
	logLevels, err := logging.GetLogLevelsForString("DEBUG")
	if err != nil {
		panic(err)
	}

	logger := logging.GetELFLogger(
		logging.Stdout,
		logLevels,
	)
	logger.SetInternalLogger(log.New(GinkgoWriter, "", 0))
	logging.SetDefaultLogger("tests", logger)
})
