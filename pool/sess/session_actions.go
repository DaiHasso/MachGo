package sess

import (
	"MachGo/base"
)

func (self Session) SaveObject(object base.Base) error {
	return saveObject(object, &self)
}

func (self Session) SaveObjects(args ...ObjectOrOption) []error {
	return saveObjects(args, &self)
}

func (self Session) GetObject(object base.Base, idValue interface{}) error {
	return getObject(object, idValue, &self)
}

func (self Session) UpdateObject(object base.Base) error {
	return updateObject(object, &self)
}

func (self Session) DeleteObject(object base.Base) error {
	return deleteObject(object, &self)
}

func (self Session) DeleteObjects(args ...ObjectOrOption) []error {
	return deleteObjects(args, &self)
}
