package dsl

type Relationship struct {
	Target,
	SelfColumn,
	TargetColumn string
}

type Relationshipable interface {
	Relationships() []Relationship
}
