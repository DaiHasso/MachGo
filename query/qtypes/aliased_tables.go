package qtypes

import (
    "fmt"
    "reflect"
    "strings"

    "github.com/pkg/errors"

    "github.com/daihasso/machgo/refl"
    "github.com/daihasso/machgo/base"
)

const tableAliasAlphabet = "abcdefghijklmnopqrstuvwxyz"+
    "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

type AliasedTables struct {
    aliasTable map[string]string
    tableAlias map[string]string
    tableType map[string]*reflect.Type
    typeTable map[reflect.Type]string
    aliasCounter int
}

func (self AliasedTables) Aliases() []string {
    var aliases []string
    for alias, _ := range self.aliasTable {
        aliases = append(aliases, alias)
    }

    return aliases
}

func (self AliasedTables) TypeForAlias(alias string) *reflect.Type {
    tableName := self.aliasTable[alias]
    return self.tableType[tableName]
}

func (self AliasedTables) TypeForTable(tableName string) *reflect.Type {
    return self.tableType[tableName]
}

func (self AliasedTables) TableForAlias(alias string) string {
    tableName := self.aliasTable[alias]
    return tableName
}

func (self AliasedTables) AliasForTable(tableName string) (string, bool) {
    alias, ok := self.tableAlias[tableName]
    return alias, ok
}

func (self AliasedTables) ObjectIsAliased(object base.Base) bool {
    tableName, err := base.BaseTable(object)
    if err != nil {
        return false
    }
    _, ok := self.tableAlias[tableName]
    return ok
}

func (self AliasedTables) ObjectAlias(object base.Base) (string, error) {
    tableName, err := base.BaseTable(object)
    if err != nil {
        return "", errors.New("Cannot determine name for object")
    }
    val, ok := self.tableAlias[tableName]
    if !ok {
        return "", errors.New("Provided Base is not aliased")
    }
    return val, nil
}

func (self AliasedTables) TypeTable(typ reflect.Type) string {
    return self.typeTable[typ]
}

func (self *AliasedTables) AddObjects(objects ...base.Base) error {
    for _, object := range objects {
        objType := reflect.TypeOf(object)
        if objType.Kind() != reflect.Ptr {
            return errors.Errorf(
                "Object provided should be *%[1]T not %[1]T", object,
            )
        }
        objType = refl.Deref(objType)
        if objType.Kind() == reflect.Ptr {
            baseType := strings.Replace(fmt.Sprintf("%T", object), "*", "", -1)
            return errors.Errorf(
                "Object provided should be *%s not %T", baseType, object,
            )
        }

        if self.aliasCounter > len(tableAliasAlphabet) {
            // TODO: Make this account for doubled aliases like aa, ab, etc.
            return errors.New(
                "There's more objects in this query than we've accounted for.",
            )
        }

        alias := string(tableAliasAlphabet[self.aliasCounter])
        tableName, err := base.BaseTable(object)
        if err != nil {
            return errors.Wrap(
                err, "Couldn't determine table name for object",
            )
        }
        self.aliasTable[alias] = tableName
        self.tableAlias[tableName] = alias

        self.tableType[tableName] = &objType
        self.typeTable[objType] = tableName

        self.aliasCounter++
    }

    return nil
}

func NewAliasedTables(objects ...base.Base) (*AliasedTables, error) {
    aliasedBases := AliasedTables{
        aliasTable: make(map[string]string, len(objects)),
        tableAlias: make(map[string]string, len(objects)),
        tableType: make(map[string]*reflect.Type, len(objects)),
        typeTable: make(map[reflect.Type]string, len(objects)),
        aliasCounter: 0,
    }

    err := aliasedBases.AddObjects(objects...)
    if err != nil {
        return nil, errors.Wrap(
            err, "Error while adding objects to AliasedTables",
        )
    }

    return &aliasedBases, nil
}
