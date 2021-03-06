package sess

import (
    "fmt"
    "sync"

    "github.com/pkg/errors"

    "github.com/daihasso/machgo/base"
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
        return err
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

func ObjectChanged(object base.Base) bool {
    hashKey, err := calculateHashKey(object)
    if err != nil {
        panic(err)
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
    tableName, err := base.BaseTable(object)
    if err != nil {
        return nil, errors.Wrap(err, "Error grabbing table name for base")
    }
    identifiers := base.GetId(object)
    hashKey := tableName
    for _, identifier := range identifiers {
        if !identifier.Exists {
            return nil, errors.New(
                "Object provided did not have an identifier",
            )
        } else if !identifier.IsSet {
            return nil, BaseIdentifierUnsetError
        }

        hashKey += fmt.Sprint(identifier.Value)
    }
    return hashKey, nil
}

func init() {
    once.Do(func() {
        globalObjectStateMutex = new(sync.RWMutex)
        savedObjects = make(map[interface{}]bool)
        objectSavedHash = make(map[interface{}]uint64)
    })
}
