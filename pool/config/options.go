package config

import (
    "github.com/daihasso/machgo/database/dbtype"

    "github.com/pkg/errors"
)

type Option func(*Config) error

func Host(host string) Option {
    return func(config *Config) error {
        config.databaseHost = host
        return nil
    }
}

func DatabaseName(name string) Option {
    return func(config *Config) error {
        config.databaseName = name
        return nil
    }
}

func Username(username string) Option {
    return func(config *Config) error {
        config.username = username
        return nil
    }
}

func Password(password string) Option {
    return func(config *Config) error {
        config.password = password
        return nil
    }
}

func Port(port int) Option {
    return func(config *Config) error {
        config.port = port
        return nil
    }
}

func PostgresType() Option {
    return func(config *Config) error {
        config.databaseType = dbtype.Postgres
        return nil
    }
}

// === Postgres specific options ===
func ensurePostgresConfig(option Option) Option {
    return func(config *Config) error {
        if config.databaseType == dbtype.UnsetDatabaseType {
            config.databaseType = dbtype.Postgres
        } else if config.databaseType != dbtype.Postgres {
            return errors.Errorf(
                "Database type is set to '%s' but it must be postgres to " +
                    "set this option.",
                config.databaseType,
            )
        }

        return option(config)
    }
}

func PostgresSSLModeOption(sslMode PostgresSSLMode) Option {
    return ensurePostgresConfig(func(config *Config) error {
        config.postgresSpecific.sslMode = sslMode

        return nil
    })
}

func PostgresExtraConnArgs(extraArgs string) Option {
    return ensurePostgresConfig(func(config *Config) error {
        config.postgresSpecific.extraConnectionArgs = extraArgs

        return nil
    })
}
