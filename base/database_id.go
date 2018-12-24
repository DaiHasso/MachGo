package base

import (
)

type DatabaseIDGenerator interface {
	DatabaseGeneratedID()
}

type DatabaseFuncIDGenerator interface {
	DatabaseIDGenerationFunc() LiteralStatement
}

type DatabaseManagedID struct {}
func (DatabaseManagedID) DatabaseGeneratedID() {}
