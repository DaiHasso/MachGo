//+build postgres !nopostgres

package database

import (
	"fmt"

	logging "github.com/daihasso/slogging"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Drivers are not directly utilized.
	"github.com/pkg/errors"
)

var postgresAddressTemplate = "user=%s password=%s host=%s port=%s " +
	"dbname=%s sslmode=require"

func getPostgresDatabase(
	username,
	password,
	serverAddress,
	port,
	dbName string,
) (*sqlx.DB, error) {
	fullAddress := fmt.Sprintf(
		postgresAddressTemplate,
		username,
		password,
		serverAddress,
		port,
		dbName,
	)

	logging.Debug("Connecting to postgres database.").
		With("database_address", fullAddress).
		Send()

	return sqlx.Open("postgres", fullAddress)
}

func PostgresConnection(connectionString string) (*sqlx.DB, error) {
	dbPool, err := sqlx.Open("postgres", connectionString)

	err = errors.Wrapf(err, "Error while opening connection to postgres")

	return dbPool, err
}
