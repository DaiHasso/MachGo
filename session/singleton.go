package session

import (
	"sync"

	"github.com/DaiHasso/MachGo/database"
)

var once sync.Once
var globalManagerMutex *sync.RWMutex

var globalManager *database.Manager

// SetGlobalManager sets the manager for all raw session initializations.
func SetGlobalManager(manager *database.Manager) {
	globalManagerMutex.Lock()
	defer globalManagerMutex.Unlock()

	globalManager = manager
}

// GetGlobalManager gets the manager for all raw session initializations.
func GetGlobalManager(manager *database.Manager) *database.Manager {
	globalManagerMutex.RLock()
	defer globalManagerMutex.RUnlock()

	return globalManager
}

func init() {
	once.Do(func() {
		globalManagerMutex = new(sync.RWMutex)
	})
}
