package dsl_test

import (
	"testing"
	"log"

	database "github.com/DaiHasso/MachGo"
	"github.com/DaiHasso/MachGo/dsl"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	logging "github.com/daihasso/slogging"
)

var (
	expectedFormat                              = `query: "%s", args: (%s)`
	object1, object2, object3, object4, object5 *testObject
	object6 *testObjectWithCreated
)

func TestDsl(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dsl Suite")
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
	logging.SetDefaultLogger("tests", &logger)

	object1 = &testObject{
		DefaultDBObject: database.DefaultDBObject{},
		table:           "testtable1",
		relationships:   make([]dsl.Relationship, 0),
	}

	object2 = &testObject{
		DefaultDBObject: database.DefaultDBObject{},
		table:           "testtable2",
		relationships: []dsl.Relationship{
			dsl.Relationship{
				Target:       object1.table,
				SelfColumn:   "foo",
				TargetColumn: "bar",
			},
		},
	}

	object3 = &testObject{
		DefaultDBObject: database.DefaultDBObject{},
		table:           "testtable3",
		relationships: []dsl.Relationship{
			dsl.Relationship{
				Target:       object1.table,
				SelfColumn:   "baz",
				TargetColumn: "test",
			},
		},
	}

	object4 = &testObject{
		DefaultDBObject: database.DefaultDBObject{},
		table:           "testtable4",
		relationships: []dsl.Relationship{
			dsl.Relationship{
				Target:       object3.table,
				SelfColumn:   "baz2",
				TargetColumn: "test2",
			},
		},
	}

	object5 = &testObject{
		DefaultDBObject: database.DefaultDBObject{},
		table:           "testtable5",
		relationships: []dsl.Relationship{
			dsl.Relationship{
				Target:       object3.table,
				SelfColumn:   "baz3",
				TargetColumn: "test3",
			},
		},
	}

	object6 = &testObjectWithCreated{
		DefaultDBObject: database.DefaultDBObject{},
		table:           "testtable3",
		relationships: []dsl.Relationship{
			dsl.Relationship{
				Target:       object1.table,
				SelfColumn:   "baz",
				TargetColumn: "test",
			},
		},
	}
})
