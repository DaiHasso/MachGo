package actrec

import (
    "reflect"

    "github.com/pkg/errors"

    "github.com/daihasso/machgo/base"
    "github.com/daihasso/machgo/query/qtypes"
    "github.com/daihasso/machgo/pool/sess"
)

type Record interface {
    Save() error
    Update() error
    Delete() error

    Column(string) qtypes.Queryable

    LinkInstance(base.Base) error
}

type DefaultRecord struct {
    instanceRef base.Base
    instanceTableName string
}

// TODO: Delete this when all the methods are implemented.
var unimplementedError = "This method hasn't been implemented"

func (self DefaultRecord) Save() error {
    err := sess.SaveObject(self.instanceRef)
    if err != nil {
        return errors.Wrap(err, "Error saving object")
    }

    return nil
}
func (self DefaultRecord) Update() error {
    // TODO: Implement.
    return errors.New(unimplementedError)
}
func (self DefaultRecord) Delete() error {
    // TODO: Implement.
    return errors.New(unimplementedError)
}
func (self DefaultRecord) Get(interface{}) error {
    // TODO: Implement.
    return errors.New(unimplementedError)
}
func (self DefaultRecord) Column(columnName string) qtypes.Queryable {
    queryable, _ := qtypes.ObjectColumn(self.instanceRef, columnName)
    return queryable
}
func (self *DefaultRecord) LinkInstance(ref base.Base) error {
    self.instanceRef = ref
    tableName, err := base.BaseTable(ref)
    if err != nil {
        return errors.Wrap(
            err, "Can't link instance; can't determine table",
        )
    }
    self.instanceTableName = tableName

    return nil
}

type ObjectActiveRecordLinker func(object base.Base) error

// FIXME: Better name pl0x.
func typeBase(arType reflect.Type) ObjectActiveRecordLinker {
    return func(object base.Base) error {
        objType := reflect.TypeOf(object)
        if objType.Kind() != reflect.Ptr ||
            objType.Elem().Kind() == reflect.Ptr {
            return errors.Errorf(
                "Object provided must be a pointer to an object not an " +
                    "object by value, pointer to a pointer or so on. " +
                    "Provided type: %T",
                object,
            )
        }
        arInType := reflect.TypeOf((*Record)(nil)).Elem()
        objVal := reflect.ValueOf(object)
        arField := objVal.Elem().FieldByName("Record")

        if !arField.IsValid() {
            return errors.New(
                "The 'Record' field does not exist on the " +
                    "object provided, did you forget to embedd the " +
                    "Record interface?",
            )
        }

        if arField.Type() != arInType {
            return errors.New(
                "The 'Record' field on the provided object " +
                    "exists but is not of the 'Record' type",
            )
        }

        newAr := reflect.New(arType)
        newArIn := newAr.Interface()
        newArInCoerced := newArIn.(Record)
        err := newArInCoerced.LinkInstance(object)
        if err != nil {
            return errors.Wrap(err, "Error linking object instance")
        }

        arField.Set(reflect.ValueOf(newArInCoerced))

        return nil
    }
}

var LinkActiveRecord = typeBase(reflect.TypeOf(
    (*DefaultRecord)(nil)).Elem(),
)

func NewActiveRecordLinker(ar Record) ObjectActiveRecordLinker {
    arType := reflect.TypeOf(ar)
    for arType.Kind() == reflect.Ptr {
        arType = arType.Elem()
    }

    return typeBase(arType)
}
