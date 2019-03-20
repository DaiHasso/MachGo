package machgo

// DefaultAttributes are a set of default attributes for an object.
type DefaultAttributes interface {
    Update()
    Init()
}
