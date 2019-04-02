// +build !nopostgres

package config

import (
    "strings"

    "github.com/pkg/errors"
    "github.com/jmoiron/sqlx"
    "github.com/daihasso/slogging"
    _ "github.com/lib/pq" // Drivers are not directly utilized.

    "github.com/daihasso/machgo/pool"
)

func postgresConnection(connectionString string) (*sqlx.DB, error) {
    dbPool, err := sqlx.Open("postgres", connectionString)

    err = errors.Wrapf(err, "Error while opening connection to postgres")

    return dbPool, err
}

func NewPostgresPool(config Config) (*pool.ConnectionPool, error) {
    connectionString := PostgresConnStringFromConfig(
        config,
    )

    cleanedConnString := connectionString
    if config.password != "" {
        cleanedConnString = strings.Replace(
            connectionString, config.password, "[omitted]", -1,
        )
    }

    logging.Debug(
        "Creating pool connection to postgres database.",
        logging.Extras{
            "connection_string": cleanedConnString,
        },
    )

    dbPool, err := postgresConnection(connectionString)
    if err != nil {
        return nil, errors.Wrap(err, "Error while creating postgres pool")
    }

    connPool := &pool.ConnectionPool{
        DB: *dbPool,
    }

    return connPool, err
}
