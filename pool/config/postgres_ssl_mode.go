package config

import (
    "github.com/go-errors/errors"
)

// PostgresSSLMode defines the modes for postgress ssl.
type PostgresSSLMode int

// This defines the modes for requiring ssl in a postgres database.
// See: https://www.postgresql.org/docs/current/libpq-ssl.html
const (
    _ PostgresSSLMode = iota
    PostgresSSLDisable
    PostgresSSLAllow
    PostgresSSLPrefer
    PostgresSSLRequire
    PostgresSSLVerifyCA
    PostgresSSLVerifyFull
)

func (self PostgresSSLMode) String() string {
    switch self {
        case PostgresSSLDisable:
        return "disable"
        case PostgresSSLAllow:
        return "allow"
        case PostgresSSLPrefer:
        return "prefer"
        case PostgresSSLRequire:
        return "require"
        case PostgresSSLVerifyCA:
        return "verify-ca"
        case PostgresSSLVerifyFull:
        return "verify-full"
        default:
        panic(errors.Errorf("Unknown SSLMode '%s'.", self))
    }
}
