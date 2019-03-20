package base

import (
    "github.com/pkg/errors"
)

func PreInsertion(object Base) error {
    if preInserter, ok := object.(PreInserter); ok {
        err := preInserter.PreInsertActions()
        if err != nil {
            return errors.Wrap(err, "Error while running PreInsertActions")
        }
    }

    return nil
}
