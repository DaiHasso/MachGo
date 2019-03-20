package sess

import (
    "github.com/daihasso/machgo/base"
)

// SaveObject saves the given object to the DB.
func (self Session) SaveObject(object base.Base) error {
    return saveObject(object, &self)
}

// SaveObjects saves the provided objects to the DB.
func (self Session) SaveObjects(args ...ObjectOrOption) []error {
    return saveObjects(args, &self)
}

// GetObject gets the object with the provided id from the DB.
func (self Session) GetObject(object base.Base, idValue interface{}) error {
    return getObject(object, idValue, &self)
}

// UpdateObject updates the object provided in the DB.
func (self Session) UpdateObject(object base.Base) error {
    return updateObject(object, &self)
}

// DeleteObject deletes the object provided from the DB.
func (self Session) DeleteObject(object base.Base) error {
    return deleteObject(object, &self)
}

// DeleteObjects deletes the objects provided from the DB.
func (self Session) DeleteObjects(args ...ObjectOrOption) []error {
    return deleteObjects(args, &self)
}
