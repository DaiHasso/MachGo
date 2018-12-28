package sess

import (
	"github.com/pkg/errors"

	"MachGo/base"
)

var saveObjectStatementTemplate = `INSERT INTO %s (%s) VALUES (%s)`

func Saved(object base.Base) bool {
	// NOTE: This requires a pointer vecause objectIsSaved and further
	//       calees assume a ptr, is this appropriate?
	saved, _ := objectIsSaved(object)
	return saved
}

func UpdateObject(object base.Base) error {
	identifier := identifierFromBase(object)
	if !identifier.exists {
		return errors.New(
			"Object provided to UpdateObject doesn't have an identifier.",
		)
	} else if !identifier.isSet {
		return errors.New(
			"Object provided to UpdateObject has an identifier but it " +
				"hasn't been set.",
		)
	}
	/*
	saved := objectIsSaved(object)
	if !saved {
		return errors.New(
			"Object provided to UpdateObject has not been saved.",
		)
	}

	if !objectChanged(object) {
		return nil
	}
	*/

	return nil
}

