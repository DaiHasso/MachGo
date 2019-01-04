package sess

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/jmoiron/sqlx"

	"MachGo/base"
)

var updateObjectStatementTemplate = `UPDATE %s SET %s WHERE %s`

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
	if !ObjectChanged(object) {
		return nil
	}

	tableName, err := base.BaseTable(object)
	if err != nil {
		return errors.Wrap(err, "Error while trying to get table name")
	}

	whereString, whereValues := updateWhere(object, identifier)

	queryParts := QueryPartsFromObject(object)

	statement := fmt.Sprintf(
		updateObjectStatementTemplate,
		tableName,
		queryParts.AsUpdate(),
		whereString,
	)

	err = Transactionized(func(tx *sqlx.Tx) error {
		var err error

		insertValues := append(queryParts.VariableValues, whereValues...)
		statement = tx.Rebind(statement)

		_, err = tx.Exec(statement, insertValues...)

		return err
	})

	if err != nil {
		return errors.Wrap(err, "Error while running update statement")
	}

	err = setObjectSaved(object)
	if err != nil {
		return errors.Wrap(err, "Error while saving object")
	}

	return nil
}
