package dsl_test

import (
	database "github.com/DaiHasso/MachGo"
	. "github.com/DaiHasso/MachGo/dsl"
	"github.com/DaiHasso/MachGo/types"
)

type testObject struct {
	database.DefaultDBObject
	table         string
	relationships []Relationship
	Id int `db:"id"`
	Name string `db:"name"`
}

type testObjectWithCreated struct {
	database.DefaultDBObject
	table         string
	relationships []Relationship
	Id int `db:"id"`
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
