package database_test

import (
    "database/sql"
	"log"
	"testing"

	logging "github.com/daihasso/slogging"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/jmoiron/sqlx"
    sqlmock "github.com/DATA-DOG/go-sqlmock"

	"MachGo"
)

var (
	db      *sql.DB
	dbx      *sqlx.DB
	mock    sqlmock.Sqlmock
)

func TestDatabase(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Database Package Suite")
}

type WeirdObject struct {}

func (*WeirdObject) GetTableName() string {
	return "weird_objects"
}

func (*WeirdObject) IsSaved() bool {
	return false
}

func (*WeirdObject) SetSaved(bool) {}
func (*WeirdObject) PreInsertActions() error { return nil }
func (*WeirdObject) PostInsertActions() error { return nil }

type FakeCompositeObject struct {
	MachGo.DefaultCompositeDBObject

	Name string `db:"name"`
	Email string `db:"email"`
	Address string `db:"address"`
}

func (*FakeCompositeObject) GetTableName() string {
	return "fake_composite_objects"
}

func (*FakeCompositeObject) GetColumnNames() []string {
	return []string{"name", "email"}
}

type FakeComplicatedObject struct {
	MachGo.DefaultDBObject

	Name string `db:"name"`
	Email string `db:"email"`
	Address string `db:"address"`
}

func (*FakeComplicatedObject) GetTableName() string {
	return "fake_complicated_objects"
}

type FakeObject struct {
	MachGo.DefaultDBObject

	Name string `db:"name"`
}

func (*FakeObject) GetTableName() string {
	return "fake_objects"
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
