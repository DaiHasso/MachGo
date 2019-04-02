package sess

import (
    "fmt"
    "database/sql"

    "github.com/pkg/errors"
    "github.com/jmoiron/sqlx"

    "github.com/daihasso/machgo/base"
)

var updateObjectStatementTemplate = `UPDATE %s SET %s WHERE %s`

func updateObject(object base.Base, session *Session) error {
    identifiers := base.GetId(object)
    for _, identifier := range identifiers {
        if !identifier.Exists {
            return errors.New(
                "Object provided to UpdateObject doesn't have an identifier.",
            )
        } else if !identifier.IsSet {
            return errors.New(
                "Object provided to UpdateObject has an identifier but it " +
                    "hasn't been set.",
            )
        }
    }
    if !ObjectChanged(object) {
        return nil
    }

    tableName, err := base.BaseTable(object)
    if err != nil {
        return errors.Wrap(err, "Error while trying to get table name")
    }

    q := updateWhere(object, identifiers)
    whereString, whereValues := q.QueryValue(nil)

    queryParts := QueryPartsFromObject(object)

    statement := fmt.Sprintf(
        updateObjectStatementTemplate,
        tableName,
        queryParts.AsUpdate(),
        whereString,
    )

    err = session.Transactionized(func(tx *sqlx.Tx) error {
        var err error

        for _, whereValue := range whereValues {
            if named, ok := whereValue.(sql.NamedArg); ok {
                queryParts.AddValue(named)
            }
        }

        statement = tx.Rebind(statement)

        _, err = tx.NamedExec(statement, queryParts.VariableValueMap)

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

// UpdateObject updates the object provided in the DB using a new session from
// the global connection pool.
func UpdateObject(object base.Base) error {
    session, err := NewSessionFromGlobal()
    if err != nil {
        return errors.Wrap(
            err, "Couldn't get session from global connection pool",
        )
    }

    return updateObject(object, session)
}
