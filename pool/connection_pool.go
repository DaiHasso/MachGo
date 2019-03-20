package pool

import (
    "github.com/daihasso/machgo/database/dbtype"

    "github.com/jmoiron/sqlx"
)

type ConnectionPool struct {
    sqlx.DB

    Type dbtype.Type
}
