package machgo

// CompositeObject is an object which has an CompositeAttribute.
type CompositeObject interface {
    Object
    CompositeKey
}
