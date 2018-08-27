package database

import (
	"fmt"

	logging "github.com/daihasso/slogging"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Drivers are not directly utilized.
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
