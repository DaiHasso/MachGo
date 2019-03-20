package sess

import (
    "fmt"
    "runtime"

    "github.com/jmoiron/sqlx"
    logging "github.com/daihasso/slogging"
    "github.com/pkg/errors"
)

func (self Session) Transactionized(
    fn func(*sqlx.Tx) error,
) (err error) {
    var tx *sqlx.Tx

    rollBack := func(tx *sqlx.Tx, oldError error) error {
        if tx != nil {
            newErr := tx.Rollback()
            if newErr != nil {
                logging.Error(
                    "Failed to rollback transaction.",
                    logging.Extras{
                        "rollback_error": newErr.Error(),
                        "initial_error": oldError.Error(),
                    },
                )
                return errors.Wrap(oldError, newErr.Error())
            }
        }

        return oldError
    }

    // Recover if something crazy happens.
    defer func() {
        if r := recover(); r != nil {
            if runtimeErr, ok := r.(runtime.Error); ok {
                panic(runtimeErr)
            }

            panicErr := errors.Wrapf(
                err, "Panic while making db transaction:\n%+v", r,
            )
            err = rollBack(tx, panicErr)
        }
    }()

    tx, err = self.Pool.Beginx()
    if err != nil {
        logging.Error("Error beginning transaction.", logging.Extras{
            "error": fmt.Sprint(err),
        })
        return rollBack(tx, err)
    }

    err = fn(tx)
    if err != nil {
        logging.Error("Error running transaction.", logging.Extras{
            "error": fmt.Sprint(err),
        })
        return rollBack(tx, err)
    }

    err = tx.Commit()
    return
}

func Transactionized(
    fn func(*sqlx.Tx) error,
) (err error) {
    session, err := NewSessionFromGlobal()
    if err != nil {
        return errors.Wrap(
            err, "Error while retrieving global ConnectionPool",
        )
    }

    return session.Transactionized(fn)
}
