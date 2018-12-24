package sess

import (
	"sync"

	logging "github.com/daihasso/slogging"
	"github.com/pkg/errors"

	"MachGo/base"
)

var once sync.Once
var globalObjectStateMutex *sync.RWMutex
var savedObjects map[interface{}]bool
var objectSavedHash map[interface{}]uint64

func setObjectSaved(object base.Base) error {
	globalObjectStateMutex.Lock()
	defer globalObjectStateMutex.Unlock()

	hashKey, err := calculateHashKey(object)
	if err != nil {
		return err
	}

	savedObjects[hashKey] = true

	hash, err := base.HashObject(object)
	if err != nil {
		logging.Warn("Error hashing object.").Send()
	}
	objectSavedHash[hashKey] = hash

	return nil
}

func objectIsSaved(object base.Base) (bool, error) {
	globalObjectStateMutex.RLock()
	defer globalObjectStateMutex.RUnlock()

	hashKey, err := calculateHashKey(object)
	if err == BaseIdentifierUnsetError {
		return false, nil
	} else if err != nil {
		return false, err
	}
	if _, ok := savedObjects[hashKey]; ok {
		return true, nil
	}

	return false, nil
}

func objectChanged(object base.Base) bool {
	hashKey, err := calculateHashKey(object)
	if err != nil {
		return true
	}

	if hash, ok := objectSavedHash[hashKey]; ok {
		newHash, err := base.HashObject(object)
		if err != nil || hash != newHash {
			return true
		}

		return false
	}

	return true
}

func calculateHashKey(object base.Base) (interface{}, error) {
	identifier := identifierFromBase(object)
	if !identifier.exists {
		return nil, errors.New("Object provided did not have an identifier.")
	} else if !identifier.isSet {
		return nil, BaseIdentifierUnsetError
	}

	tableName, err := base.BaseTable(object)
	if err != nil {
		return nil, err
	}
	hashKey := [2]interface{}{identifier.value, tableName}
	return hashKey, nil
}

func init() {
	once.Do(func() {
		globalObjectStateMutex = new(sync.RWMutex)
		savedObjects = make(map[interface{}]bool)
		objectSavedHash = make(map[interface{}]uint64)
	})
}
