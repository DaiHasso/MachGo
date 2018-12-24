package base_test

import (
    . "github.com/onsi/ginkgo"
    . "github.com/onsi/gomega"

	. "MachGo/base"
)

type ARObjectTest struct {
	ActiveRecorder
}

var _ = Describe("ActiveRecordLinker", func() {
	It("Should be able to link an active record to an instance.", func() {
		object := ARObjectTest{}
		err := LinkActiveRecord(&object)
		Expect(err).ToNot(HaveOccurred())
		err = object.Save()
		Expect(err).To(HaveOccurred())
	})

	It("Should error when an object is passed by value", func() {
		object := ARObjectTest{}
		err := LinkActiveRecord(object)
		Expect(err).To(HaveOccurred())
	})

	It("Should error when an object is a pointer to a pointer", func() {
		object := ARObjectTest{}
		objPtr := &object
		err := LinkActiveRecord(&objPtr)
		Expect(err).To(HaveOccurred())
	})
})
