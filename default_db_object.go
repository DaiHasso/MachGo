package MachGo

import (
	"fmt"
)

// DefaultDBObject is a basic object with standard implementations.
type DefaultDBObject struct {
	jsonExportedDefaultAttributes
	DefaultDiffable
	ID *IntID `json:"id" db:"id"`

	saved bool
}

// IsSaved will return a bool encapsulating wether the object has been
// saved or not.
func (bo *DefaultDBObject) IsSaved() bool {
	return bo.saved
}

// SetSaved sets the saved status for the object.
func (bo *DefaultDBObject) SetSaved(saved bool) {
	bo.saved = saved
}

// SetID sets the ID.
func (bo *DefaultDBObject) SetID(id ID) error {
	intID, ok := id.(*IntID)
	if !ok {
		return fmt.Errorf("Couldn't convert ID '%s' to int64", id)
	}

	bo.ID = intID

	return nil
}

// GetID retrieves the ID.
func (bo DefaultDBObject) GetID() ID {
	return bo.ID
}

// IDIsSet will check if the ID is set and return true if it has been set.
func (bo DefaultDBObject) IDIsSet() bool {
	return bo.ID == nil
}

// GetIDColumn returns the column the ID lives in.
func (bo DefaultDBObject) GetIDColumn() string {
	return "id"
}

// NewID returns nil because it can't generate a new ID without seeing existing
// IDs.
func (bo *DefaultDBObject) NewID() ID {
	return nil
}

// PreInsertActions will initialize the default attributes.
func (bo *DefaultDBObject) PreInsertActions() (err error) {
	if bo.IsSaved() {
		bo.jsonExportedDefaultAttributes.update()
	} else {
		bo.jsonExportedDefaultAttributes.init()
	}

	return
}

// PostInsertActions is a NOOP for default UUID object.
func (bo *DefaultDBObject) PostInsertActions() (err error) { return }
