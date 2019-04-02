package config

import (
    "github.com/daihasso/machgo/database/dbtype"
    "github.com/daihasso/machgo/pool"

    "time"

    "github.com/pkg/errors"
)

type Config struct {
    databaseHost,
    databaseName,
    username,
    password string
    port int
    databaseType dbtype.Type
    databaseTimeout time.Duration
    postgresSpecific postgresConfig
}

type postgresConfig struct {
    sslMode PostgresSSLMode
    extraConnectionArgs string
}

var DefaultPostgresConfig = Config{
    databaseHost: "localhost",
    databaseName: "postgres",
    username: "root",
    password: "",
    port: 5432,
    databaseType: dbtype.Postgres,
    postgresSpecific: postgresConfig{
        // NOTE: I'd rather this was prefer but lib/pq doesn't currently
        //       support allow or prefer. See:
        //       https://github.com/lib/pq/issues/776
        sslMode: PostgresSSLRequire,
    },
}

func poolFromConfig(
    config Config,
    handler func(Config) (*pool.ConnectionPool, error),
    options ...Option,
) (*pool.ConnectionPool, error) {
    for i, option := range options {
        err := option(&config)
        if err != nil {
            return nil, errors.Wrapf(
                err, "Error while adding option %d:", i,
            )
        }
    }
    connPool, err := handler(config)

    if err != nil {
        return nil, errors.Wrapf(
            err, "Error while %s creating pool", config.databaseType,
        )
    }

    return connPool, nil
}

func PostgresPool(options ...Option) (*pool.ConnectionPool, error) {
    config := DefaultPostgresConfig

    return poolFromConfig(config, NewPostgresPool, options...)
}
