package machgo

// IDAttribute is an interface for doing common ID operations.
type IDAttribute interface {
    SetID(ID) error
    GetID() ID
    IDIsSet() bool
    NewID() ID
    GetIDColumn() string
}
