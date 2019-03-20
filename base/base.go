package base

type Base interface {}

type PreInserter interface {
    PreInsertActions() error
}

type PostInserter interface {
    PostInsertActions() error
}

type Saveable interface {
    Saved() bool
    Save()
}

type TraditionalID interface {
    SetID(interface{}) error
    ID() (interface{}, bool)
}

type IDColumner interface {
    IDColumn() string
}

type IDGenerator interface {
    NewID() interface{}
}
