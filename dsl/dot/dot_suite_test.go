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
	MachGo.DefaultDBObject
	table         string
	relationships []MachGo.Relationship
	Name string `db:"name"`
}

func (self *testObject) GetTableName() string {
	return self.table
}

func (self *testObject) Relationships() []MachGo.Relationship {
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
		DefaultDBObject: MachGo.DefaultDBObject{},
		table:           "testtable1",
		relationships:   make([]MachGo.Relationship, 0),
	}

	object2 = &testObject{
		DefaultDBObject: MachGo.DefaultDBObject{},
		table:           "testtable2",
	}

	object3 = &testObject{
		DefaultDBObject: MachGo.DefaultDBObject{},
		table:           "testtable3",
	}

	object4 = &testObject{
		DefaultDBObject: MachGo.DefaultDBObject{},
		table:           "testtable4",
	}

	object5 = &testObject{
		DefaultDBObject: MachGo.DefaultDBObject{},
		table:           "testtable5",
	}

	object2.relationships = []MachGo.Relationship{
		MachGo.Relationship{
			SelfObject: object2,
			TargetObject: object1,
			SelfColumn: "foo",
			TargetColumn: "bar",
		},
	}
	object3.relationships = []MachGo.Relationship{
		MachGo.Relationship{
			SelfObject: object3,
			TargetObject: object1,
			SelfColumn: "baz",
			TargetColumn: "test",
		},
	}
	object4.relationships = []MachGo.Relationship{
		MachGo.Relationship{
			SelfObject: object4,
			TargetObject: object3,
			SelfColumn:   "baz2",
			TargetColumn: "test2",
		},
	}
	object5.relationships = []MachGo.Relationship{
		MachGo.Relationship{
			SelfObject: object5,
			TargetObject: object3,
			SelfColumn:   "baz3",
			TargetColumn: "test3",
		},
	}
})
