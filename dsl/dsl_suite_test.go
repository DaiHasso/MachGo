package dsl_test

import (
	"testing"
	"log"

	database "github.com/DaiHasso/MachGo"
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
