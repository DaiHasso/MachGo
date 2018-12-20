package sess

import (
	"MachGo"
	"MachGo/pool"
	"MachGo/dsl"
	"MachGo/database"
)

// Session is a helper wrapper for a connection pool that has common helper
// functions.
type Session struct {
	Pool *pool.ConnectionPool
}

func (self Session) Query(objects ...MachGo.Object) *dsl.QuerySequence {
	qs := dsl.NewQuerySequence()
	if len(objects) > 0 {
		qs.Join(objects...)
	}
	qs.SetPool(self.Pool)
	return qs
}

func (self Session) Manager() *database.Manager {
	return database.NewManagerFromPool(self.Pool)
}

func NewSessionFromGlobal() (*Session, error) {
	connPool, err := pool.GlobalConnectionPool()
	return &Session{connPool}, err
}

func NewSession() (*Session, error) {
	return NewSessionFromGlobal()
}

func NewSessionFromPool(connPool *pool.ConnectionPool) *Session {
	return &Session{connPool}
}

func Query(objects ...MachGo.Object) *dsl.QuerySequence {
	qs := dsl.NewQuerySequence()
	if len(objects) > 0 {
		qs.Join(objects...)
	}
	qs.SetPool(self.Pool)
	return qs
}
