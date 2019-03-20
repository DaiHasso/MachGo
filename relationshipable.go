package machgo

// Relationship is a representation of how two objects join together.
type Relationship struct {
    SelfObject,
    TargetObject Object
    SelfColumn,
    TargetColumn string
}

// Invert takes the relationship and swaps the self with the target.
// This essentially has the effect of changing:
//     foo.bar=baz.fizz
// Into:
//     baz.fizz=foo.bar
func (self Relationship) Invert() *Relationship {
    return &Relationship {
        SelfObject: self.TargetObject,
        TargetObject: self.SelfObject,
        SelfColumn: self.TargetColumn,
        TargetColumn: self.SelfColumn,
    }
}

func (self Relationship) SelfTable() string {
    return self.SelfObject.GetTableName()
}

func (self Relationship) TargetTable() string {
    return self.TargetObject.GetTableName()
}

// Relationshipable is a type that has at least on relationship.
type Relationshipable interface {
    Relationships() []Relationship
}
