package sess

import (
    "github.com/daihasso/machgo/base"
    "github.com/daihasso/machgo/pool"
    "github.com/daihasso/machgo/query"
    "github.com/daihasso/machgo/database"
)

// Session is a helper wrapper for a connection pool that has common helper
// functions.
type Session struct {
    Pool *pool.ConnectionPool
}

func (self Session) Query(objects ...base.Base) *query.Query {
    q := query.NewQuery(self.Pool)
    if len(objects) > 0 {
        q.Join(objects...)
    }
    return q
}

func (self Session) Manager() (*database.Manager, error) {
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

func Query(objects ...base.Base) *query.Query {
    connPool, err := pool.GlobalConnectionPool()
    if err != nil {
        return nil
    }
    q := query.NewQuery(connPool)
    if len(objects) > 0 {
        q.Join(objects...)
    }
    return q
}
