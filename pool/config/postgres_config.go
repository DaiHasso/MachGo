package config

import (
    "fmt"
    "math"
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
