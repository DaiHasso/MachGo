package pool

import (
    "database/sql"
   
    "github.com/daihasso/machgo/pool/dbtype"

    "github.com/jmoiron/sqlx"
)

type ConnectionPool struct {
    sqlx.DB

    Type dbtype.Type
}

func ConnectionPoolFromDb(db *sql.DB, dbType dbtype.Type) *ConnectionPool {
    dbx := sqlx.NewDb(db, string(dbType))

    return &ConnectionPool{
        DB: *dbx,
        Type: dbType,
    }
}

func ConnectionPoolFromDbx(db *sqlx.DB) *ConnectionPool {
    return &ConnectionPool{
        DB: *db,
        Type: dbtype.Type(db.DriverName()),
    }
}
