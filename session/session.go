package session

import (
	"MachGo"
	"MachGo/dsl"
	"MachGo/database"
)

type Session struct {
	manager *database.Manager
}

func New() *Session {
	globalManagerMutex.RLock()
	return &Session{
		manager: globalManager,
	}
}

func (self Session) Query(objects ...MachGo.Object) *dsl.QuerySequence {
	qs := dsl.NewQuerySequence()
	if len(objects) > 0 {
		qs.Join(objects...)
	}
	qs.SetManager(self.manager)
	return qs
}

func (self Session) Close() {
	globalManagerMutex.RUnlock()
}
