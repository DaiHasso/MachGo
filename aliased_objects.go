package MachGo

import (
	"errors"
	"reflect"

	"github.com/DaiHasso/MachGo/refl"
)

const tableAliasAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

type AliasedObjects struct {
    aliasTable map[string]string
    tableAlias map[string]string
    tableType map[string]*reflect.Type
    typeAlias map[reflect.Type]string
	objAliasCounter int
}

func (self AliasedObjects) TypeForAlias(alias string) *reflect.Type {
	tableName := self.aliasTable[alias]
	return self.tableType[tableName]
}

func (self AliasedObjects) TableForAlias(alias string) string {
	tableName := self.aliasTable[alias]
	return tableName
}

func (self AliasedObjects) ObjectIsAliased(object Object) bool {
	_, ok := self.tableAlias[object.GetTableName()]
	return ok
}

func (self AliasedObjects) ObjectAlias(
	object Object,
) (string, error) {
	val, ok := self.tableAlias[object.GetTableName()]
	if !ok {
		return "", errors.New("Provided Object is not aliased.")
	}
	return val, nil
}

// NewAliasedObjectsFromExisting takes an oldfashion map-style aliased object
// and creates a new AliasedObjects struct.
func NewAliasedObjectsFromExisting(
	aliasedObjectMap map[string]Object,
) (*AliasedObjects, error) {
	aliasedObjects := AliasedObjects{
		aliasTable: make(map[string]string, len(aliasedObjectMap)),
		tableAlias: make(map[string]string, len(aliasedObjectMap)),
		tableType: make(map[string]*reflect.Type, len(aliasedObjectMap)),
		objAliasCounter: len(aliasedObjectMap),
	}
	for alias, object := range aliasedObjectMap {
		tableName := object.GetTableName()
		aliasedObjects.aliasTable[alias] = tableName
		aliasedObjects.tableAlias[tableName] = alias

		objType := refl.Deref(reflect.TypeOf(object))
		aliasedObjects.tableType[tableName] = &objType

	}

	return &aliasedObjects, nil
}

func NewAliasedObjects(objects ...Object) (*AliasedObjects, error) {
	aliasedObjects := AliasedObjects{
		aliasTable: make(map[string]string, len(objects)),
		tableAlias: make(map[string]string, len(objects)),
		tableType: make(map[string]*reflect.Type, len(objects)),
		objAliasCounter: 0,
	}

	for _, object := range objects {
		if aliasedObjects.objAliasCounter > len(tableAliasAlphabet) {
			// TODO: Make this account for doubled aliases like aa, ab, etc.
			return nil, errors.New(
				"There's more objects in this query than we've accounted for.",
			)
		}

		alias := string(tableAliasAlphabet[aliasedObjects.objAliasCounter])
		tableName := object.GetTableName()
		aliasedObjects.aliasTable[alias] = tableName
		aliasedObjects.tableAlias[tableName] = alias

		objType := refl.Deref(reflect.TypeOf(object))
		aliasedObjects.tableType[tableName] = &objType

		aliasedObjects.objAliasCounter++
	}

	return &aliasedObjects, nil
}
