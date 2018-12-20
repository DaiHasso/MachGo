package config

import (
	"MachGo/database"
	"MachGo/pool"

	"fmt"
	"math"
	"strings"

	logging "github.com/daihasso/slogging"
	"github.com/pkg/errors"
)

var postgresAddressTemplate = "dbname='%s' user='%s' host='%s' port='%d' " +
	"sslmode='%s'"
var postgresPasswordTemplate = "password='%s'"
var postgresConnectionTimeout = "connect_timeout='%d'"

func PostgresConnStringFromConfig(config Config) string {
	connectionString := fmt.Sprintf(
		postgresAddressTemplate,
		config.databaseName,
		config.username,
		config.databaseHost,
		config.port,
		config.postgresSpecific.sslMode,
	)

	if config.password != "" {
		connectionString += " " + fmt.Sprintf(
			postgresPasswordTemplate,
			config.password,
		)
	}

	if config.databaseTimeout != 0 {
		connectionString += " " + fmt.Sprintf(
			postgresConnectionTimeout,
			int64(math.Round(config.databaseTimeout.Seconds())),
		)
	}

	if config.postgresSpecific.extraConnectionArgs != "" {
		connectionString += " " + config.postgresSpecific.extraConnectionArgs
	}

	return connectionString
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

	logging.Debug("Creating pool connection to postgres database.").With(
		"connection_string", cleanedConnString,
	).Send()

	dbPool, err := database.PostgresConnection(connectionString)
	if err != nil {
		return nil, errors.Wrap(err, "Error while creating postgres pool")
	}

	connPool := &pool.ConnectionPool{
		DB: *dbPool,
	}

	return connPool, err
}
