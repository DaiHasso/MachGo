package sess

import (
	"database/sql"
	"fmt"

	logging "github.com/daihasso/slogging"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"MachGo/base"
	"MachGo/pool"
	"MachGo/database/dbtype"
)

func SaveObject(object base.Base) error {
	var (
		err error
		columnFilters []func(string) bool
		databaseManagedId bool
	)

	identifier, err := initIdentifier(object)
	if err != nil {
		return err
	}

	saved, err := objectIsSaved(object)
	if err != nil {
		return err
	}
	if saved {
		return UpdateObject(object)
	}

	idColumn := objectIdColumn(object)
	if !identifier.isSet {
		if _, ok := object.(base.DatabaseIDGenerator); ok {
			databaseManagedId = true
			removeID := func(columnName string) bool {
				if columnName == idColumn {
					return true
				}

				return false
			}
			columnFilters = append(columnFilters, removeID)
		} else {
			return errors.New(
				"Object has no dentifier set.",
			)
		}
	}

	tableName, err := base.BaseTable(object)
	if err != nil {
		return errors.Wrap(err, "Error while trying to get table name")
	}

	err = base.PreInsertion(object)
	if err != nil {
		return err
	}

	// NOTE: Great pains are taken here to maintain a consistent ordering
	//       (sorted on name) for the columns and values. The chief reason
	//       for this is for testing reproducibility. It might be
	//       marginally faster to not do this, so if things prove to be a
	//       little slow in practice it might be worth revisiting this
	//       approach or at least having it toggle-able in some way.

	var (
		namesString, bindVarsString string
		sortedNamedValues []interface{}
	)
	empty := true
	processSortedNamedValues(
		object, func(columnName string, namedArg *sql.NamedArg) {
			for _, filter := range columnFilters {
				if filter(columnName) {
					return
				}
			}

			if empty {
				empty = false
			} else {
				namesString += ", "
				bindVarsString += ", "
			}

			namesString += columnName
			bindVarsString += "@" + columnName
			sortedNamedValues = append(sortedNamedValues, *namedArg)
		},
	)

	query := fmt.Sprintf(
		saveObjectStatementTemplate,
		tableName,
		namesString,
		bindVarsString,
	)

	logging.Debug("Running SaveObject statement.").
		With("statement", query).
		With("object_type", fmt.Sprintf("%T", object)).
		With("values", fmt.Sprintf("%#+v", sortedNamedValues)).
		Send()

	var dbType dbtype.Type
	if databaseManagedId {
		connPool, _ := pool.GlobalConnectionPool()
		dbType = connPool.Type
	}

	err = Transactionized(func(tx *sqlx.Tx) error {
		var err error

		query = tx.Rebind(query)

		if databaseManagedId {
			if dbType == dbtype.Postgres {
				if dbType == dbtype.Postgres {
					query = fmt.Sprintf("%s RETURNING %s", query, idColumn)
				}
				row := tx.QueryRowx(query, sortedNamedValues...)

				err := row.StructScan(object)
				if err != nil {
					return errors.Wrap(
						err, "Error while reading returned id from database",
					)
				}
			} else if dbType == dbtype.Mysql {
				result, err := tx.Exec(query, sortedNamedValues...)
				if err != nil {
					return errors.Wrap(
						err, "Error executing insert",
					)
				}

				id, err := result.LastInsertId()
				if err != nil {
					return errors.Wrap(
						err, "Couldn't get returned id from database",
					)
				}

				err = setIdentifierOnBase(object, id)
				if err != nil {
					return errors.Wrap(err, "Error setting new id on object")
				}
			} else {
				return errors.Errorf(
					"Unsupported db type '%s' for database " +
					"managed ID",
					dbType,
				)
			}
		} else {
			_, err = tx.Exec(query, sortedNamedValues...)
			if err != nil {
				return err
			}
		}

		if postInserter, ok := object.(base.PostInserter); ok {
			err = postInserter.PostInsertActions()
			if err != nil {
				return errors.Wrap(
					err, "Error while running PostInsertActions",
				)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	err = setObjectSaved(object)
	if err != nil {
		return err
	}

	return nil
}
