package dsl_test

import (
	database "github.com/DaiHasso/MachGo"
	. "github.com/DaiHasso/MachGo/dsl"
)

type testObject struct {
	database.DefaultDBObject
	table         string
	relationships []Relationship
	Id int `db:"id"`
}

func (self *testObject) GetTableName() string {
	return self.table
}

func (self *testObject) Relationships() []Relationship {
	return self.relationships
}
