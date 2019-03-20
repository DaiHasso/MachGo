package sess

import (
    "database/sql"
    "fmt"

    logging "github.com/daihasso/slogging"
    "github.com/jmoiron/sqlx"
    "github.com/pkg/errors"

    "github.com/daihasso/machgo/base"
    "github.com/daihasso/machgo/database/dbtype"
)

var saveObjectStatementTemplate = `INSERT INTO %s %s`

func saveObjects(
    args []ObjectOrOption, session *Session,
) []error {
    // NOTE: Originally this went through all the objects and constructed one
    //       insert with multiple values but that presented problems with
    //       reading back IDs from the DB. Maybe there's a better way but this
    //       seems like the best for now.
    objects, options := separateAndApply(args)

    var allErrors []error
    for i, object := range objects {
        err := saveObject(object, session)
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

func saveObject(object base.Base, session *Session) error {
    var (
        err error
        columnFilters []ColumnFilter
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
            removeID := func(columnName string, _ *sql.NamedArg) bool {
                return columnName == idColumn
            }
            columnFilters = append(columnFilters, removeID)
        } else {
            return errors.New(
                "Object has no identifier set and does not have an ID" +
                " generator.",
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

    queryParts := QueryPartsFromObject(object, columnFilters...)

    query := fmt.Sprintf(
        saveObjectStatementTemplate, tableName, queryParts.AsInsert(),
    )

    logging.Debug("Running SaveObject statement.", logging.Extras{
        "statement": query,
        "object_type": fmt.Sprintf("%T", object),
        "values": fmt.Sprintf("%#+v", queryParts.VariableValues),
    })

    err = doInsertion(
        session,
        object,
        query,
        idColumn,
        queryParts.VariableValues,
        databaseManagedId,
    )
    if err != nil {
        return err
    }

    err = setObjectSaved(object)
    if err != nil {
        return err
    }

    return nil
}

func insertActionWithPostInserter(
    session *Session,
    object base.Base,
    statement string,
    insertAction func(tx *sqlx.Tx) error,
) error {
    return session.Transactionized(func(tx *sqlx.Tx) error {
        var err error

        statement = tx.Rebind(statement)

        err = insertAction(tx)
        if err != nil {
            return err
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
}

func basicInsert(
    session *Session,
    object base.Base,
    statement, idColumn string,
    insertValues []interface{},
) error {
    return insertActionWithPostInserter(
        session,
        object,
        statement,
        func(tx *sqlx.Tx) error {
            _, err := tx.Exec(statement, insertValues...)
            return err
        },
    )
}

func insertReturningId(
    session *Session,
    object base.Base,
    statement, idColumn string,
    dbType dbtype.Type,
    insertValues []interface{},
) error {
    return insertActionWithPostInserter(
        session,
        object,
        statement,
        func(tx *sqlx.Tx) error {
            return insertReadingId(
                object, statement, idColumn, dbType, insertValues, tx,
            )
        },
    )
}

func insertReadingId(
    object base.Base,
    statement, idColumn string,
    dbType dbtype.Type,
    insertValues []interface{},
    tx *sqlx.Tx,
) error {
    if dbType == dbtype.Postgres {
        if dbType == dbtype.Postgres {
            statement = fmt.Sprintf("%s RETURNING %s", statement, idColumn)
        }
        row := tx.QueryRowx(statement, insertValues...)

        err := row.StructScan(object)
        if err != nil {
            return errors.Wrap(
                err, "Error while reading returned id from database",
            )
        }
    } else if dbType == dbtype.Mysql {
        result, err := tx.Exec(statement, insertValues...)
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
    return nil
}

func doInsertion(
    session *Session,
    object base.Base,
    statement, idColumn string,
    insertValues []interface{},
    databaseManagedId bool,
) error {
    if databaseManagedId {
        dbType := session.Pool.Type
        return insertReturningId(
            session, object, statement, idColumn, dbType, insertValues,
        )
    }

    return basicInsert(
        session, object, statement, idColumn, insertValues,
    )
}

// SaveObject saves the given object to the DB new session from the global
// connection pool.
func SaveObject(object base.Base) error {
    session, err := NewSessionFromGlobal()
    if err != nil {
        return errors.Wrap(
            err, "Couldn't get session from global connection pool",
        )
    }

    return saveObject(object, session)
}

// SaveObjects saves the provided objects to the DB using a new session from
// the global connection pool.
func SaveObjects(args ...ObjectOrOption) []error {
    session, err := NewSessionFromGlobal()
    if err != nil {
        return []error{
            errors.Wrap(
                err, "Couldn't get session from global connection pool",
            ),
        }
    }
    return saveObjects(args, session)
}
