package actrec

import (
    "reflect"

    "github.com/pkg/errors"

    "github.com/daihasso/machgo/base"
    //"github.com/daihasso/machgo/query"
    "github.com/daihasso/machgo/query/qtypes"
)

type RecordAccessor interface {
    //Get(interface{}) error
    Find(...qtypes.Queryable) (qtypes.QueryResults, error)
    All() (qtypes.QueryResults, error)
    //Where(...qtypes.Queryable) *query.Query
    //FindBy(string, interface{}) base.Base
}

type DefaultRecordAccessor struct {
    instanceType reflect.Type
    instanceTableName string
}

func (self DefaultRecordAccessor) Find(
    ...qtypes.Queryable,
) (qtypes.QueryResults, error) {
    // TODO: Implement.
    return qtypes.QueryResults{}, errors.New(unimplementedError)
}
func (self DefaultRecordAccessor) All() (qtypes.QueryResults, error) {
    // TODO: Implement.
    return qtypes.QueryResults{}, errors.New(unimplementedError)
}

func NewRecordAccessor(object base.Base) (RecordAccessor, error) {
    objType := reflect.TypeOf(object)
    for objType.Kind() == reflect.Ptr {
        objType = objType.Elem()
    }
    tableName, err := base.BaseTable(object)
    if err != nil {
        return nil, errors.Wrap(
            err, "Can't link instance; can't determine table",
        )
    }

    return &DefaultRecordAccessor{
        instanceType: objType,
        instanceTableName: tableName,
    }, nil
}
