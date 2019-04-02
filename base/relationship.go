package base

import (
    "reflect"

    "github.com/pkg/errors"

    "github.com/daihasso/machgo/refl"
)

type typeTableColumn struct {
    typ reflect.Type
    column,
    table string
}

// Relationship is a representation of how two objects join together.
type Relationship struct {
    selfInfo,
    targetInfo typeTableColumn
}

// Invert takes the relationship and swaps the self with the target.
// This essentially has the effect of changing:
//     foo.bar=baz.fizz
// Into:
//     baz.fizz=foo.bar
func (self Relationship) Invert() *Relationship {
    return &Relationship {
        selfInfo: self.targetInfo,
        targetInfo: self.selfInfo,
    }
}

func (self Relationship) Tables() (string, string) {
    return self.selfInfo.table, self.targetInfo.table
}

func (self Relationship) Columns() (string, string) {
    return self.selfInfo.column, self.targetInfo.column
}

func (self Relationship) Types() (reflect.Type, reflect.Type) {
    return self.selfInfo.typ, self.targetInfo.typ
}

// Relationshipable is a type that has at least on relationship.
type Relationshipable interface {
    Relationships() []Relationship
}

func NewRelationship(
    self Base, selfColumn string, target Base, targetColumn string,
) (*Relationship, error) {
    selfTable, err := BaseTable(self)
    if err != nil {
        return nil, errors.Wrap(
            err, "Error determining table for self in relationship",
        )
    }
    targetTable, err := BaseTable(target)
    if err != nil {
        return nil, errors.Wrap(
            err, "Error determining table for target in relationship",
        )
    }

    targetInfo := typeTableColumn{
        typ: refl.Deref(reflect.TypeOf(target)),
        column: targetColumn,
        table: targetTable,
    }

    selfInfo := typeTableColumn{
        typ: refl.Deref(reflect.TypeOf(self)),
        column: selfColumn,
        table: selfTable,
    }

    return &Relationship{
        selfInfo: selfInfo,
        targetInfo: targetInfo,
    }, nil
}

func MustRelationship(
    self Base, selfColumn string, target Base, targetColumn string,
) Relationship {
    selfTable, err := BaseTable(self)
    if err != nil {
        panic(errors.Wrap(
            err, "Error determining table for self in relationship",
        ))
    }
    targetTable, err := BaseTable(target)
    if err != nil {
        panic(errors.Wrap(
            err, "Error determining table for target in relationship",
        ))
    }

    targetInfo := typeTableColumn{
        typ: refl.Deref(reflect.TypeOf(target)),
        column: targetColumn,
        table: targetTable,
    }

    selfInfo := typeTableColumn{
        typ: refl.Deref(reflect.TypeOf(self)),
        column: selfColumn,
        table: selfTable,
    }

    return Relationship{
        selfInfo: selfInfo,
        targetInfo: targetInfo,
    }
}
