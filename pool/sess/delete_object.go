package sess

import (
    "database/sql"
    "fmt"

    "github.com/pkg/errors"
    "github.com/jmoiron/sqlx"

    "github.com/daihasso/machgo/base"
)

var deleteObjectStatementTemplate = `DELETE FROM %s WHERE %s`

func deleteObject(object base.Base, session *Session) error {
    identifier := identifierFromBase(object)
    if !identifier.exists {
        return errors.New(
            "Object provided to DeleteObject doesn't have an identifier.",
        )
    } else if !identifier.isSet {
        return errors.New(
            "Object provided to DeleteObject has an identifier but it " +
                "hasn't been set.",
        )
    }
    idColumn := objectIdColumn(object)

    tableName, err := base.BaseTable(object)
    if err != nil {
        return errors.Wrap(err, "Error while trying to get table name")
    }

    // TODO: Support composites.
    idValue := sql.Named(idColumn, identifier.value)

    whereClause := fmt.Sprintf("%s = @%s", idColumn, idColumn)

    statement := fmt.Sprintf(
        deleteObjectStatementTemplate,
        tableName,
        whereClause,
    )

    err = session.Transactionized(func(tx *sqlx.Tx) error {
        var err error

        statement = tx.Rebind(statement)

        insertValues := []interface{}{idValue}
        _, err = tx.Exec(statement, insertValues...)

        return err
    })
    if err != nil {
        return errors.Wrap(err, "Error while running delete statement")
    }

    return nil
}

func deleteObjects(
    args []ObjectOrOption, session *Session,
) []error {
    // NOTE: Originally this went through all the objects and constructed one
    //       insert with multiple values but that presented problems with
    //       reading back IDs from the DB. Maybe there's a better way but this
    //       seems like the best for now.

    objects, options := separateAndApply(args)

    var allErrors []error
    for i, object := range objects {
        err := deleteObject(object, session)
        if err != nil {
            allErrors = append(
                allErrors,
                errors.Wrapf(err, "Error while saving object #%d", i+1),
            )
            if options.stopOnFailure {
                return allErrors
            }
        }
    }

    return allErrors
}

// DeleteObject deletes the object provided from the DB using a new session
// from the global connection pool.
func DeleteObject(object base.Base) error {
    session, err := NewSessionFromGlobal()
    if err != nil {
        return errors.Wrap(
            err, "Couldn't get session from global connection pool",
        )
    }

    return deleteObject(object, session)
}

// DeleteObjects deletes the objects provided from the DB using a new session
// from the global connection pool.
func DeleteObjects(args ...ObjectOrOption) []error {
    session, err := NewSessionFromGlobal()
    if err != nil {
        return []error{
            errors.Wrap(
                err, "Couldn't get session from global connection pool",
            ),
        }
    }

    return deleteObjects(args, session)
}
