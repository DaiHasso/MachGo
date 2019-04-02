// +build nopostgres

package config

import (
    "github.com/pkg/errors"
)

func NewPostgresPool(config Config) (*pool.ConnectionPool, error) {
    return nil, errors.New(
        "Postgres not enabled, re-build without nopostgres flag to use " +
            "postgres",
    )
}
