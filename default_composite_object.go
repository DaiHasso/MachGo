package MachGo

import "errors"

// DefaultCompositeDBObject is a basic object with standard implementations.
type DefaultCompositeDBObject struct {
	JSONExportedDefaultAttributes
	DefaultDiffable

	saved bool
}

// IsSaved will return a bool encapsulating wether the object has been
// saved or not.
func (bo DefaultCompositeDBObject) IsSaved() bool {
	return bo.saved
}

// SetSaved modifies the saved status of the object.
func (bo *DefaultCompositeDBObject) SetSaved(saved bool) {
	bo.saved = saved
}

// PreInsertActions will initialize the default attributes.
func (bo *DefaultCompositeDBObject) PreInsertActions() (err error) {
	if bo.IsSaved() {
		bo.JSONExportedDefaultAttributes.Update()
	} else {
		bo.JSONExportedDefaultAttributes.Init()
	}

	return
}

// PostInsertActions is a NOOP for default UUID object.
func (bo *DefaultCompositeDBObject) PostInsertActions() (err error) { return }

// GetColumnNames retrieves the database column names for the composite key.
func (bo *DefaultCompositeDBObject) GetColumnNames() []string {
	panic(errors.New("you haven't overriden your column names methods"))
}

// SetColumnNames retrieves the database column names for the composite key.
func (bo *DefaultCompositeDBObject) SetColumnNames(columns []string) {
	panic(errors.New("you haven't overriden your column names methods"))
}
