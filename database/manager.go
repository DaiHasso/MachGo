package database

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"

	logging "github.com/daihasso/slogging"

	"github.com/jmoiron/sqlx"
	"github.com/DaiHasso/MachGo/refl"
	. "github.com/DaiHasso/MachGo"
)

// Manager is a database wrapper with a few helpful tools
// for working with objects.
type Manager struct {
	*sqlx.DB

	databaseType Type
}

// GetObject gets an object by its ID.
func (m *Manager) GetObject(obj Object, id ID) error {
	queryTemplate := `SELECT * FROM %s WHERE id=?`

	query := fmt.Sprintf(
		queryTemplate,
		obj.GetTableName(),
	)

	query = m.Rebind(query)

	logging.Debug("Find object.").
		With("query", query).
		With("id", id).
		Send()

	err := m.Get(obj, query, id)
	if err != nil {
		translatedError := translateDBError(err)
		logging.Error("Failed to get object").
			With("error_message", translatedError.Error()).
			Send()
		return translatedError
	}

	obj.SetSaved(true)

	return err
}

// FindObject finds an object by intelligently reading attributes of
// provided object.
func (m *Manager) FindObject(obj IDObject) error {
	queryTemplate := `SELECT * FROM %s %s`

	nameTagValueInterfaces := refl.GetFieldsByTagWithTagValues(obj, "db")

	var whereClause []byte
	var variableValues []interface{}

	if obj.IDIsSet() {
		if len(whereClause) != 0 {
			whereClause = append(whereClause, []byte(" AND ")...)
		} else {
			whereClause = append(whereClause, []byte("WHERE ")...)
		}
		whereClause = append(whereClause, []byte("id = ?")...)
		variableValues = append(
			variableValues,
			obj.GetID(),
		)
	} else {
		for _, tagValueInterface := range nameTagValueInterfaces {
			if tagValueInterface.IsUnset() {
				continue
			}

			if len(whereClause) != 0 {
				whereClause = append(whereClause, []byte(" AND ")...)
			} else {
				whereClause = append(whereClause, []byte("WHERE ")...)
			}

			variableNameSQL := []byte(tagValueInterface.TagValue)
			whereClause = append(whereClause, variableNameSQL...)
			whereClause = append(whereClause, []byte(" = ?")...)
			variableValues = append(
				variableValues,
				tagValueInterface.Interface,
			)
		}
	}

	query := fmt.Sprintf(
		queryTemplate,
		obj.GetTableName(),
		string(whereClause),
	)

	query = m.Rebind(query)

	logging.Debug("Find object.").
		With("query", query).
		With("values", variableValues).
		Send()

	err := m.Get(obj, query, variableValues...)
	if err != nil {
		translatedError := translateDBError(err)
		logging.Error("Failed to find object.").
			With("error_message", translatedError.Error()).
			Send()
		return translatedError
	}

	obj.SetSaved(true)

	return err
}

// FindObjects will find all objects matching queryObject parameter and return
// an interface representing a pointer to a slice of pointers to the type of
// the queryObject with the values found.
// Ex: queryObject is a Post then the return value will be a *[]*Post.
func (m *Manager) FindObjects(
	queryObject Object,
) (interface{}, error) {
	queryTemplate := `SELECT * FROM %s %s`
	whereTemplate := `WHERE %s`

	nameTagValueInterfaces := refl.GetFieldsByTagWithTagValues(
		queryObject,
		"db",
	)

	var whereClause string
	var whereValues []interface{}
	if result, ok := queryObject.(IDObject); ok {
		newWhereClause, newWhereValues := buildIDWhereClause(
			nameTagValueInterfaces,
			result.GetID(),
		)
		whereClause = string(newWhereClause)
		whereValues = append(whereValues, newWhereValues...)
	} else if result, ok := queryObject.(CompositeObject); ok {
		for _, tagValueInterface := range nameTagValueInterfaces {
			newWhereValues, newWhereClause, err := buildCompositeWhereClause(
				result,
				tagValueInterface,
				whereClause,
				false,
			)
			if err != nil {
				return nil, err
			}
			whereClause = newWhereClause
			whereValues = append(whereValues, newWhereValues...)
		}
		whereClause = fmt.Sprintf(whereTemplate, whereClause)
	} else {
		return nil, errors.New(
			"provided object is neither an IDObject nor a " +
				"CompositeObject, I don't know how to handle this.",
		)
	}

	query := fmt.Sprintf(
		queryTemplate,
		queryObject.GetTableName(),
		whereClause,
	)

	query = m.Rebind(query)

	logging.Debug("Find objects.").
		With("query_object", queryObject).
		With("query", query).
		With("values", whereValues).
		Send()

	iface := refl.GetInterfaceSlice(queryObject)

	err := m.Select(iface, query, whereValues...)
	if err != nil {
		logging.Warn("Could not perform select for FindObjects.").
			With("error", err).
			Send()
		return nil, err
	}

	logging.Debug("Finished query.").
		With("results", fmt.Sprintf("%#v", iface)).
		Send()

	return iface, nil
}

// SaveObject automatically saves an object to the database.
// TODO: Consider converting resolved name into snake case and using as
//       database table if database table unset.
func (m *Manager) SaveObject(obj Object) error {
	if obj.IsSaved() {
		return m.UpdateObject(obj)
	}

	action := func(obj Object, tx *sqlx.Tx) error {
		var err error

		err = obj.PreInsertActions()
		if err != nil {
			return err
		}
		queryTemplate := `INSERT INTO %s (%s) VALUES (%s)`

		nameTagValueInterfaces := refl.GetFieldsByTagWithTagValues(obj, "db")

		var variableNames, variableBindVars []byte
		var variableValues []interface{}
		var needsIDFromDB = false

		if idObj, ok := obj.(IDObject); ok {
			idColumn, idValue := buildInitID(idObj)

			if idValue == nil {
				needsIDFromDB = true
			} else {
				variableNames = append(variableNames, []byte(idColumn)...)
				variableBindVars = append(variableBindVars, '?')
				variableValues = append(variableValues, idValue)
			}
		}

		for name, tagValueInterface := range nameTagValueInterfaces {
			if len(variableNames) != 0 {
				variableNames = append(variableNames, []byte{',', ' '}...)
				variableBindVars = append(
					variableBindVars,
					[]byte{',', ' '}...,
				)
			}

			variableNameSQL := []byte(tagValueInterface.TagValue)
			variableNames = append(variableNames, variableNameSQL...)
			variableBindVars = append(variableBindVars, '?')
			variableValues = append(
				variableValues,
				tagValueInterface.Interface,
			)
			if diffable, ok := obj.(Diffable); ok {
				diffable.SetLastSavedValue(name, tagValueInterface.Interface)
			}
		}

		query := fmt.Sprintf(
			queryTemplate,
			obj.GetTableName(),
			string(variableNames),
			string(variableBindVars),
		)

		query = m.Rebind(query)

		logging.Debug("Running save object query.").
			With("query", query).
			With("object_type", refl.GetInterfaceName(obj)).
			With("values", variableValues).
			Send()

		if needsIDFromDB {
			if idObj, ok := obj.(IDObject); ok {
				err = insertAndSetID(
					idObj,
					query,
					variableValues,
					tx,
					m.databaseType,
				)
			}
		} else {
			_, err = tx.Exec(query, variableValues...)
		}
		if err != nil {
			return err
		}

		obj.SetSaved(true)

		err = obj.PostInsertActions()
		return err
	}

	return m.objectTransaction(action, obj)
}

// UpdateObject will take an object and write an appropriate update
// statement to update it's values.
func (m *Manager) UpdateObject(obj Object) error {
	if !obj.IsSaved() {
		return ErrObjectNotSaved
	}

	action := func(obj Object, tx *sqlx.Tx) error {
		queryTemplate := `UPDATE %s SET %s WHERE %s`
		var whereClause string
		var compositeObj CompositeObject
		var variableSetStatements []byte
		var variableValues []interface{}
		var whereValues []interface{}

		if result, ok := obj.(IDObject); ok {
			whereClause = "id=?"
			variableValues = append(variableValues, result.GetID())
		} else if result, ok := obj.(CompositeObject); ok {
			compositeObj = result
		} else {
			panic(errors.New(
				"provided object is neither an IDObject nor a " +
					"CompositeObject, I don't know how to handle this.",
			))
		}

		err := obj.PreInsertActions()
		if err != nil {
			panic(err)
		}

		nameTagValueInterfaces := refl.GetFieldsByTagWithTagValues(obj, "db")

		diffable, isDiffable := obj.(Diffable)

		for name, tagValueInterface := range nameTagValueInterfaces {
			if isDiffable {
				if diffable.GetLastSavedValue(name) == tagValueInterface.Interface {
					continue
				} else {
					diffable.SetLastSavedValue(name, tagValueInterface.Interface)
				}
			}

			if compositeObj != nil {
				var (
					newWhereValues []interface{}
					newWhereClause string
				)
				newWhereValues, newWhereClause, err = buildCompositeWhereClause(
					compositeObj,
					tagValueInterface,
					whereClause,
					true,
				)
				if err != nil {
					return err
				}
				whereClause = newWhereClause
				whereValues = append(whereValues, newWhereValues...)
			}

			if len(variableSetStatements) != 0 {
				variableSetStatements = append(variableSetStatements, []byte{',', ' '}...)
			}

			variableSetSQL := []byte(
				tagValueInterface.TagValue + "=?",
			)

			variableSetStatements = append(
				variableSetStatements,
				variableSetSQL...,
			)

			variableValues = append(
				variableValues,
				tagValueInterface.Interface,
			)
		}

		variableValues = append(
			whereValues,
			variableValues...,
		)

		query := fmt.Sprintf(
			queryTemplate,
			obj.GetTableName(),
			string(variableSetStatements),
			whereClause,
		)

		query = m.Rebind(query)

		logging.Debug("Updating object.").
			With("query", query).
			Send()

		_, err = tx.Exec(query, variableValues...)
		if err != nil {
			return translateDBError(err)
		}

		return nil
	}

	return m.objectTransaction(action, obj)
}

// DeleteObject will delete an object.
func (m *Manager) DeleteObject(obj Object) error {
	action := func(obj Object, tx *sqlx.Tx) error {
		queryTemplate := "DELETE FROM %s %s"
		var whereClause string
		var whereValues []interface{}

		nameTagValueInterfaces := refl.GetFieldsByTagWithTagValues(obj, "db")

		if result, ok := obj.(IDObject); ok {
			newWhereClause, newWhereValues := buildIDWhereClause(
				nameTagValueInterfaces,
				result.GetID(),
			)
			whereClause = string(newWhereClause)
			whereValues = append(whereValues, newWhereValues...)
		} else if result, ok := obj.(CompositeObject); ok {
			for _, tagValueInterface := range nameTagValueInterfaces {
				newWhereValues, newWhereClause, err := buildCompositeWhereClause(
					result,
					tagValueInterface,
					whereClause,
					true,
				)
				if err != nil {
					return err
				}
				whereClause = newWhereClause
				whereValues = append(whereValues, newWhereValues...)
			}
		} else {
			panic(errors.New(
				"provided object is neither an IDObject nor a " +
					"CompositeObject, I don't know how to handle this.",
			))
		}

		query := fmt.Sprintf(
			queryTemplate,
			obj.GetTableName(),
			whereClause,
		)

		query = m.Rebind(query)

		logging.Debug("Making delete query.").
			With("query", query).
			Send()

		_, err := tx.Exec(query, whereValues...)
		if err != nil {
			return translateDBError(err)
		}

		return nil
	}

	return m.objectTransaction(action, obj)
}

// CreateTableForObject tries to create a table for an object if it doesn't
// exist. It will read sqlType tags if available or try to infer based on
// type reflection what type of column to create.
// TODO: Handle non-mysql databases.
func (m *Manager) CreateTableForObject(obj Object) error {
	// TODO: Actually do something.

	return nil
}

// TODO: Is this really necessary?
func (m *Manager) objectTransaction(
	fn func(Object, *sqlx.Tx) error,
	obj Object,
) error {
	wrapped := func(tx *sqlx.Tx) error {
		return fn(obj, tx)
	}

	return m.Transactionized(wrapped)
}

// Transactionized wraps the provided function in a transaction that
// automatically translates errors and rolls back any transactions in case of
// failure.
func (m *Manager) Transactionized(
	fn func(*sqlx.Tx) error,
) (err error) {
	var tx *sqlx.Tx

	rollBack := func(tx *sqlx.Tx, oldError error) error {
		if tx != nil {
			newErr := tx.Rollback()
			if newErr != nil {
				logging.Error("Failed to rollback transaction.").
					With("rollback_error", newErr.Error()).
					With("initial_error", oldError.Error()).
					Send()
				return newErr
			}
		}

		return oldError
	}

	// Recover if something crazy happens.
	defer func() {
		if r := recover(); r != nil {
			err = translateDBError(rollBack(tx, fmt.Errorf("%s", r)))
		}
	}()

	tx, err = m.Beginx()
	if err != nil {
		logging.Error("Error beginning transaction.").
			With("error", fmt.Sprint(err)).
			Send()
		return translateDBError(rollBack(tx, err))
	}

	err = fn(tx)
	if err != nil {
		logging.Error("Error running transaction.").
			With("error", fmt.Sprint(err)).
			Send()
		return translateDBError(rollBack(tx, err))
	}

	err = tx.Commit()
	return
}

// GetDatabaseManager will init and retrieve a mysql database.
func GetDatabaseManager(
	databaseType Type,
	username,
	password,
	serverAddress,
	port,
	dbName string,
) (*Manager, error) {
	var err error
	var db *sqlx.DB

	switch databaseType {
	case Mysql:
		db, err = getMysqlDatabase(
			username,
			password,
			serverAddress,
			port,
			dbName,
		)
	case Postgres:
		db, err = getPostgresDatabase(
			username,
			password,
			serverAddress,
			port,
			dbName,
		)
	default:
		logging.Error("Unsupported database type.").
			With("database_type", databaseType).
			Send()
		err = errors.New("Unsupported database type")
	}
	if err != nil {
		return nil, err
	}

	manager := &Manager{db, databaseType}

	return manager, nil
}

// NewManagerFromExisting will create a new Manager from an existing database.
func NewManagerFromExisting(
	databaseType Type,
	existing *sql.DB,
	databaseName string,
) (*Manager, error) {
	sqlxDB := sqlx.NewDb(existing, databaseName)
	manager := &Manager{sqlxDB, databaseType}

	return manager, nil
}

func insertAndSetID(
	obj IDObject,
	query string,
	variableValues []interface{},
	tx *sqlx.Tx,
	databaseType Type,
) error {
	switch databaseType {
	case Mysql:
		result, err := tx.Exec(query, variableValues...)
		if err != nil {
			logging.Error("Couldn't exec query.").
				With("error", err).
				Send()
			return err
		}

		id, err := result.LastInsertId()
		if err != nil {
			logging.Error("Couldn't get last insert ID.").
				With("error", err).
				Send()
			return err
		}

		intID := IntID{id}

		err = obj.SetID(&intID)
		if err != nil {
			logging.Error("Error while trying to set ID.").
				With("error", err).Send()
		}
	case Postgres:
		query = fmt.Sprintf("%s%s", query, " RETURNING id")
		row := tx.QueryRowx(query, variableValues...)

		err := row.StructScan(obj)
		if err != nil {
			logging.Error("Couldn't scan row into object.").
				With("error", err).
				Send()
			return err
		}
	default:
		return fmt.Errorf("Unsupported DB type %s", databaseType.String())
	}

	return nil
}

func buildIDWhereClause(
	nameTagValueInterfaces map[string]refl.TagValueInterface,
	id driver.Valuer,
) ([]byte, []interface{}) {
	var whereClause []byte
	var variableValues []interface{}

	var idValue driver.Value
	var err error
	if id != nil {
		idValue, err = id.Value()
	}
	if err == nil && idValue != nil {
		if len(whereClause) != 0 {
			whereClause = append(whereClause, []byte(" AND ")...)
		} else {
			whereClause = append(whereClause, []byte("WHERE ")...)
		}
		whereClause = append(whereClause, []byte("id = ?")...)
		variableValues = append(
			variableValues,
			id,
		)
	} else {
		for _, tagValueInterface := range nameTagValueInterfaces {
			if tagValueInterface.IsUnset() {
				continue
			}

			if len(whereClause) != 0 {
				whereClause = append(whereClause, []byte(" AND ")...)
			} else {
				whereClause = append(whereClause, []byte("WHERE ")...)
			}

			variableNameSQL := []byte(tagValueInterface.TagValue)
			whereClause = append(whereClause, variableNameSQL...)
			whereClause = append(whereClause, []byte(" = ?")...)
			variableValues = append(
				variableValues,
				tagValueInterface.Interface,
			)
		}
	}

	return whereClause, variableValues
}

func buildInitID(id IDAttribute) (string, ID) {
	newID := id.NewID()
	err := id.SetID(newID)
	if err != nil {
		panic(err)
	}
	return id.GetIDColumn(), id.GetID()
}

func buildCompositeWhereClause(
	compositeObj CompositeObject,
	tagValueInterface refl.TagValueInterface,
	existingWhereClause string,
	strict bool,
) ([]interface{}, string, error) {
	newVariableValues := make([]interface{}, 0)
	newWhereClause := existingWhereClause

	// TODO: Maybe columns should be a map?
	for _, column := range compositeObj.GetColumnNames() {
		if column == tagValueInterface.TagValue {
			if tagValueInterface.IsUnset() {
				if strict {
					logging.Warn(
						"A composite object was queried and required all "+
							"columns to be filled but some columns were "+
							"unset.",
					).With("unset_column", column).With(
						"composite_columns",
						compositeObj.GetColumnNames(),
					).Send()

					return nil, "", errors.New(
						"action requires all composite keys to be present",
					)
				}
				continue
			}

			if len(newWhereClause) != 0 {
				newWhereClause += ` AND `
			}

			newWhereClause += fmt.Sprintf(
				`%s=?`,
				tagValueInterface.TagValue,
			)

			newVariableValues = append(
				newVariableValues,
				tagValueInterface.Interface,
			)
		}
	}

	return newVariableValues, newWhereClause, nil
}
