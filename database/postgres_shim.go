//+build nopostgres,!postgres

package database

import (
    "github.com/pkg/errors"
)

func getPostgresDatabase(
    username,
    password,
    serverAddress,
    port,
    dbName string,
) (*sqlx.DB, error) {
    return nil, errors.New(
        "machgo built without postgres capability, cannot create postgres " +
            "connection.",
    )
}

func PostgresConnection(connectionString string) (*sqlx.DB, error) {
    return nil, errors.New(
        "machgo built without postgres capability, cannot create postgres " +
            "connection.",
    )
}
