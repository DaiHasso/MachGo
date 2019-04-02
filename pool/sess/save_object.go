package sess

import (
    "database/sql"
    "fmt"
    "strings"

    "github.com/daihasso/slogging"
    "github.com/jmoiron/sqlx"
    "github.com/pkg/errors"

    "github.com/daihasso/machgo/base"
    "github.com/daihasso/machgo/pool/dbtype"
)

var saveObjectStatementTemplate = `INSERT INTO %s %s`

func saveObjects(
    args []ObjectsOrOptions, session *Session,
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

    identifiers, err := base.InitializeId(object)
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

    var idColumns []string
    for _, identifier := range identifiers {
        if !identifier.IsSet {
            idColumn := identifier.Column
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

        idColumns = append(idColumns, identifier.Column)
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
        "values": fmt.Sprintf("%#+v", queryParts.VariableValueMap),
    })

    err = doInsertion(
        session,
        object,
        query,
        idColumns,
        queryParts.VariableValueMap,
        databaseManagedId,
    )
    if err != nil {
        return errors.Wrap(
            err,
            "Error saving object",
        )
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
    err := session.Transactionized(func(tx *sqlx.Tx) error {
        var err error

        statement = tx.Rebind(statement)

        err = insertAction(tx)
        if err != nil {
            return err
        }

        return nil
    })
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
}

func basicInsert(
    session *Session,
    object base.Base,
    statement string,
    idColumn []string,
    insertValues map[string]interface{},
) error {
    return insertActionWithPostInserter(
        session,
        object,
        statement,
        func(tx *sqlx.Tx) error {
            statement = tx.Rebind(statement)
            _, err := tx.NamedExec(statement, insertValues)
            return err
        },
    )
}

func insertReturningId(
    session *Session,
    object base.Base,
    statement string,
    idColumn []string,
    dbType dbtype.Type,
    insertValues map[string]interface{},
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
    statement string,
    idColumns []string,
    dbType dbtype.Type,
    insertValues map[string]interface{},
    tx *sqlx.Tx,
) error {
    if dbType == dbtype.Postgres {
        if dbType == dbtype.Postgres {
            columns := strings.Join(idColumns, ", ")
            statement = fmt.Sprintf("%s RETURNING %s", statement, columns)
        }
        row, err := tx.NamedQuery(statement, insertValues)
        if err != nil {
            return errors.Wrap(
                err, "Error while preparing query",
            )
        }

        err = row.StructScan(object)
        if err != nil {
            return errors.Wrap(
                err, "Error while reading returned id from database",
            )
        }
    } else if dbType == dbtype.Mysql {
        // FIXME: Actually implement this.
        return errors.New(
            "Database managed composites are hard in Mysql so I'm punting " +
                "on this",
        )
        /*
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
        */
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
    statement string,
    idColumn []string,
    insertValues map[string]interface{},
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
func SaveObjects(args ...ObjectsOrOptions) []error {
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
