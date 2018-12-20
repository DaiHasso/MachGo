package pool

import (
	"MachGo/database/dbtype"

	"github.com/jmoiron/sqlx"
)

type ConnectionPool struct {
	sqlx.DB

	Type dbtype.Type
}
