package dbtype

import (
    "fmt"
    "strings"
)

// Type describes a type of database.
//go:generate stringer -type=Type
type Type int

// Definitions of different database types.
const (
    UnsetDatabaseType Type = iota
    Mysql
    Postgres
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
        return 0, fmt.Errorf("Unknown database type '%s'", typeString)
    }
}

func (self Type) String() string {
    switch(self) {
    case Mysql:
        return "mysql"
    case Postgres:
        return "postgres"
    default:
        return "unknown"
    }
}
