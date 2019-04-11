package sess

import (
    "fmt"
    "reflect"

    "github.com/jmoiron/sqlx"
    "github.com/pkg/errors"

    "github.com/daihasso/machgo/base"
)

var getObjectStatementTemplate = `SELECT * FROM %s WHERE %s`

func getObject(
    target base.Base, idValue interface{}, session *Session,
) error {
    var err error

    identifiers := base.GetId(target)
    if len(identifiers) > 1 {
        return errors.New("Can't use get with a composite object")
    }
    identifier := identifiers[0]
    if !identifier.Exists {
        return errors.New(
            "Object provided to GetObject doesn't have an identifier.",
        )
    } else if identifier.IsSet {
        return errors.New(
            "Object provided to GetObject has an identifier set, it should " +
                "be a new instance with no identifier.",
        )
    }

    if identifier.Value != nil &&
        reflect.TypeOf(identifier.Value) != reflect.TypeOf(idValue) {
        return errors.Errorf(
            "Type of provided id (%T) does not match identifier type for " +
                "object (%T).",
            idValue,
            identifier.Value,
        )
    }

    idColumn := objectIdColumn(target)

    tableName, err := base.BaseTable(target)
    if err != nil {
        return errors.Wrap(err, "Error while trying to get table name")
    }

    whereClause := fmt.Sprintf("%s = :%s", idColumn, idColumn)

    statement := fmt.Sprintf(
        getObjectStatementTemplate, tableName, whereClause,
    )

    values := map[string]interface{}{
        idColumn: idValue,
    }

    err = session.Transactionized(func(tx *sqlx.Tx) error {
        var err error
        statement = tx.Rebind(statement)

        rows, err := tx.NamedQuery(statement, values)
        if err != nil {
            return errors.Wrap(
                err, "Error while reading data from DB",
            )
        }
        defer rows.Close()

        if !rows.Next() {
            return errors.New(
                "No results from DB for object with provided id.",
            )
        }

        err = rows.StructScan(target)
        if err != nil {
            return errors.Wrap(
                err, "Error while reading data from DB into struct",
            )
        }

        return nil
    })
    if err != nil {
        return err
    }

    err = setObjectSaved(target)
    if err != nil {
        return err
    }

    return nil
}

// GetObject gets the object with the provided id from the DB using a new
// session from the global connection pool.
func GetObject(object base.Base, idValue interface{}) error {
    session, err := NewSessionFromGlobal()
    if err != nil {
        return errors.Wrap(
            err, "Couldn't get session from global connection pool",
        )
    }

    return getObject(object, idValue, session)
}
