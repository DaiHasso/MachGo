package MachGo

// Diffable allows something to track changes to the item before saving to
// optimize inserts.
type Diffable interface {
	GetLastSavedValue(string) interface{}
	SetLastSavedValue(string, interface{})
}

// DefaultDiffable is the basic standard implementation of a diffable.
type DefaultDiffable struct {
	lastSavedValues map[string]interface{}
}

// SetLastSavedValue will set a given value in the last values map.
func (dd *DefaultDiffable) SetLastSavedValue(name string, val interface{}) {
	if dd.lastSavedValues == nil {
		dd.lastSavedValues = make(map[string]interface{})
	}
	dd.lastSavedValues[name] = val
}

// GetLastSavedValue will get a given value from the last values map.
func (dd DefaultDiffable) GetLastSavedValue(name string) interface{} {
	return dd.lastSavedValues[name]
}
