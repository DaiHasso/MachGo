package pool

import (
	"sync"

	"github.com/pkg/errors"
)

var once sync.Once
var globalConnectionPool *ConnectionPool
var globalConnectionPoolMutex *sync.RWMutex

func SetGlobalConnectionPool(pool *ConnectionPool) {
	globalConnectionPoolMutex.Lock()
	defer globalConnectionPoolMutex.Unlock()
	globalConnectionPool = pool
}

func GlobalConnectionPool() (*ConnectionPool, error) {
	globalConnectionPoolMutex.RLock()
	defer globalConnectionPoolMutex.RUnlock()
	if globalConnectionPool == nil {
		return nil, errors.New(
			"GetGlobalConnectionPool called without global ConnectionPool " +
				"being established.",
		)
	}
	return globalConnectionPool, nil
}

func init() {
	once.Do(func() {
		globalConnectionPoolMutex = new(sync.RWMutex)
	})
}
