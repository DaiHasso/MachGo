package dbtype

import (
    "strings"

    "github.com/pkg/errors"
)

// Type describes a type of database.
type Type string

// Definitions of different database types.
const (
    UnsetDatabaseType Type = ""
    Mysql Type = "mysql"
    Postgres Type = "postgres"
)

// TypeFromString will take a string representation of a database type
// and return a Type.
func TypeFromString(typeString string) (Type, error) {
    switch strings.ToLower(typeString) {
    case "mysql":
        return Mysql, nil
    case "postgres", "pgsql", "psql":
        return Postgres, nil
    default:
        return "", errors.Errorf("Unknown database type '%s'", typeString)
    }
}
