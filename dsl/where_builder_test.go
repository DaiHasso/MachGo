package dsl_test

import (
	"daihasso/mew/tests"
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/DaiHasso/MachGo/dsl"
)

func whereNoSequence(t *testing.T) {
	whereString := dsl.NewWhere().ObjectColumn(object1, "foo").Eq().Const(1).
		String()

	tests.AssertEqual(t, whereString, "testtable1.foo = 1")

	t.Log(whereString)
}

func whereComplexNoSequence(t *testing.T) {
	whereString := dsl.NewWhere().ObjectColumn(object1, "foo").Eq().Const(1).And().
		ObjectColumn(object1, "bar").LessEq().ObjectColumn(object1, "baz").Or().
		ObjectColumn(object2, "id").Greater().ObjectColumn(object1, "modified").
		String()

	tests.AssertEqual(
		t,
		whereString,
		"testtable1.foo = 1 AND testtable1.bar <= testtable1.baz OR "+
			"testtable2.id > testtable1.modified",
	)
	t.Log(whereString)
}

func whereSubConditionNoSequence(t *testing.T) {
	whereString := dsl.NewWhere().ObjectColumn(object1, "foo").Eq().Const(1).And().
		SubCond(
			dsl.NewWhere().ObjectColumn(object1, "bar").LessEq().
				ObjectColumn(object1, "baz").Or().
				ObjectColumn(object2, "id").Greater().
				ObjectColumn(object1, "modified"),
		).
		String()
	tests.AssertEqual(
		t,
		whereString,
		"testtable1.foo = 1 AND (testtable1.bar <= testtable1.baz OR "+
			"testtable2.id > testtable1.modified)",
	)
	t.Log(whereString)
}

func whereDoubleSubConditionNoSequence(t *testing.T) {
	whereString := dsl.NewWhere().ObjectColumn(object1, "foo").Eq().Const(1).And().
		SubCond(
			dsl.NewWhere().ObjectColumn(object1, "bar").LessEq().
				ObjectColumn(object1, "baz").Or().
				SubCond(
					dsl.NewWhere().ObjectColumn(object2, "id").Greater().
						ObjectColumn(object1, "modified"),
				)).
		String()
	tests.AssertEqual(
		t,
		whereString,
		"testtable1.foo = 1 AND (testtable1.bar <= testtable1.baz OR "+
			"(testtable2.id > testtable1.modified))",
	)
	t.Log(whereString)
}

var _ = Describe("WhereBuilder", func() {
	It("Should generate a string for a simple where", func() {
		whereString := dsl.NewWhere().ObjectColumn(object1, "foo").Eq().
			Const(1).String()

		fmt.Fprintf(GinkgoWriter, whereString)

		Expect(whereString).To(Equal("testtable1.foo = 1"))
	})
})
