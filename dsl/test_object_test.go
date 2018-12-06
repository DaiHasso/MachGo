package dsl_test

import (
	. "MachGo"
	"MachGo/types"
)

type testObject struct {
	DefaultDBObject
	table         string
	relationships []Relationship
	Name string `db:"name"`
}

type testObjectWithCreated struct {
	DefaultDBObject
	table         string
	relationships []Relationship
	Name string `db:"name"`
	Created types.Timestamp
}

func (self *testObject) GetTableName() string {
	return self.table
}

func (self *testObject) Relationships() []Relationship {
	return self.relationships
}

func (self *testObjectWithCreated) GetTableName() string {
	return self.table
}

func (self *testObjectWithCreated) Relationships() []Relationship {
	return self.relationships
}
