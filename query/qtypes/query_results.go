package qtypes

import (
    "reflect"

    "github.com/jmoiron/sqlx"
    "github.com/pkg/errors"

    "github.com/daihasso/machgo/refl"
    "github.com/daihasso/machgo/base"
)

// BaseSlicePointer is a pointer to a slice of pointers to bases like:
//   `*[]*MyObject`
// It is represented by an interface so that it can take in a slice of any
// custom object you've implemented in your project.
type BaseSlicePointer interface{}

// QueryResults represents a set of results which generate QueryResult per row
// and have convinience functions for batch reading.
type QueryResults struct {
    tx *sqlx.Tx
    rows *sqlx.Rows
    nextResult *QueryResult
    aliasedTables *AliasedTables
    typeBSFieldMap map[reflect.Type]*refl.GroupedFieldsWithBS
    lastError error
    columns []string
    columnAliasFields []ColumnAliasField
    aliasesInSelect map[string]bool

    closed bool
}

func columnsToFieldNames(
    columnNames []string,
    typeBSFieldMap map[reflect.Type]*refl.GroupedFieldsWithBS,
    aliasedTables *AliasedTables,
) ([]ColumnAliasField, map[string]bool, error) {
    columnAliasFields := make([]ColumnAliasField, len(columnNames))
    aliasesInSelect := make(map[string]bool, len(columnNames))
    for i, column := range columnNames {
        columnAlias, ok := ColumnAliasFromString(column)
        if !ok {
            return nil, nil, errors.Errorf(
                "Unexpected column in result: '%s'",
                column,
            )
        }

        aliasesInSelect[columnAlias.TableAlias] = true

        objType := aliasedTables.TypeForAlias(columnAlias.TableAlias)
        tagValBSFields := *typeBSFieldMap[*objType]

        var fieldName string
        if bsField, ok := tagValBSFields[columnAlias.ColumnName]; ok {
            fieldName = bsField.Name()
        } else {
            fieldName = SnakeToUpperCamel(columnAlias.ColumnName)
        }

        columnAliasField := ColumnAliasField{
            ColumnAlias: *columnAlias,
            FieldName: fieldName,
        }

        columnAliasFields[i] = columnAliasField
    }

    return columnAliasFields, aliasesInSelect, nil
}

// Next retrieves the next result from the result set and indicates if there
// are more. This mimics the sql.Rows pattern and sets errors which can be
// accessed via the Err() function.
func (self *QueryResults) Next() bool {
    hasRows := self.rows.Next()
    err := self.rows.Err()
    if err != nil {
        self.lastError = err
        return false
    }
    if !hasRows {
        return false
    }

    if self.columns == nil {
        columnNames, err := self.rows.Columns()
        if err != nil {
            self.lastError = err
            return false
        }

        self.columns = columnNames

        columnAliasFields, aliasesInSelect, err := columnsToFieldNames(
            self.columns,
            self.typeBSFieldMap,
            self.aliasedTables,
        )
        if err != nil {
            self.lastError = err
            return false
        }

        self.columnAliasFields = columnAliasFields
        self.aliasesInSelect = aliasesInSelect
    }

    self.nextResult, err = NewQueryResult(
        self.rows,
        self.aliasedTables,
        self.columnAliasFields,
    )
    if err != nil {
        self.lastError = err
        return false
    }
    self.nextResult.closeAfterWrite = false

    return true
}

// GetResult returns the current result if it has been prepped with Next().
func (self *QueryResults) GetResult() *QueryResult {
    return self.nextResult
}

// Err returns any errors if they have occured.
func (self *QueryResults) Err() error {
    return self.lastError
}

// Close closes this QueryResults' rows and commits/rollsback the transaction
// it wraps. This can be safely called multiple times.
func (self *QueryResults) Close() error {
    if !self.closed {
        self.rows.Close()
        err := self.tx.Commit()
        if err != nil {
            rollErr := self.tx.Rollback()
            if rollErr != nil {
                return errors.Wrapf(
                    err,
                    "Error rolling back transaction caused by error while " +
                        "commiting '%s'",
                    rollErr.Error(),
                )
            }
        }

        self.closed = true
    }

    return nil
}

// WriteN writes `count` items to the provided slices automatically determining
// what data to write to which slices. This operation does not close the
// transaction.
func (self *QueryResults) WriteN(
    count int, objectSlices ...BaseSlicePointer,
) (retErr error) {
    return self.write(count, false, objectSlices)
}

func (self *QueryResults) write(
    count int, closeAfter bool, objectSlices []BaseSlicePointer,
) (retErr error) {
    defer func() {
        r := recover()
        if retErr != nil || r != nil {
            if retErr == nil {
                retErr = errors.Errorf(
                    "Panic while writing results '%#+v'",
                    r,
                )
            }
            self.rows.Close()
            newErr := self.tx.Rollback()
            if newErr != nil {
                retErr = errors.Wrapf(
                    newErr,
                    "Error rolling back transaction in response to '%s'",
                    retErr.Error(),
                )
            }
        } else if closeAfter {
            err := self.Close()
            if err != nil {
                retErr = err
            }
        }
    }()

    elemTypes := make([]reflect.Type, len(objectSlices))
    for i, objectSlice := range objectSlices {
        typ := reflect.TypeOf(objectSlice)
        if typ.Kind() != reflect.Ptr {
            return errors.Errorf(
                "Type of arg #%d must be pointer to slice not %T",
                i,
                objectSlice,
            )
        }
        if elemTyp := typ.Elem(); elemTyp.Kind() != reflect.Slice {
            return errors.Errorf(
                "Type of arg #%d must be pointer to slice type not a " +
                    "pointer to a %s",
                i,
                elemTyp,
            )
        }
        objElemType, err := refl.ElementTypeFromSlice(objectSlice)
        if err != nil {
            return errors.WithStack(err)
        }

        objElemTypeDeref := objElemType
        if objElemType.Kind() == reflect.Ptr {
            objElemTypeDeref = objElemType.Elem()
        }

        if _, ok := self.typeBSFieldMap[objElemTypeDeref]; !ok {
            return errors.Errorf(
                "No results match type '%s'.", objElemTypeDeref,
            )
        }


        elemTypes[i] = objElemType
    }

    checkCount := func(total int) bool {
        if count >= 0 && total > count {
            return false
        }

        return true
    }

    for i := 0; self.Next() && checkCount(count); i++ {
        elemValues := make([]base.Base, len(elemTypes))
        for i, elemTypePtr := range elemTypes {
            elemType := elemTypePtr.Elem()
            elemVal := reflect.New(elemType).Interface()
            elemAlias, err := self.aliasedTables.ObjectAlias(
                elemVal.(base.Base),
            )
            if err != nil {
                return errors.WithStack(err)
            }
            if _, ok := self.aliasesInSelect[elemAlias]; !ok {
                return errors.Errorf(
                    "Can't write slice of type '%s', object is not in select.",
                    elemType,
                )
            }

            elemValues[i] = elemVal
        }
        nextResult := self.GetResult()
        err := nextResult.WriteTo(elemValues...)
        if err != nil {
            return errors.WithStack(err)
        }

        for i, elemVal := range elemValues {
            objSlice := objectSlices[i]
            sliceVal := reflect.ValueOf(objSlice)
            newStruct := reflect.Append(
                sliceVal.Elem(),
                reflect.ValueOf(elemVal),
            )

            sliceVal.Elem().Set(newStruct)
        }
    }

    return nil
}

// WriteAllTo writes all results to the provided slices, automatically
// determining which match which. This operation closes the transaction.
func (self *QueryResults) WriteAllTo(objectSlices ...BaseSlicePointer) error {
    return self.write(-1, true, objectSlices)
}

// NewQueryResults returns a new QueryResults from a finished query with
// pending rows.
func NewQueryResults(
    tx *sqlx.Tx,
    rows *sqlx.Rows,
    aliasedTables *AliasedTables,
    typeBSFieldMap map[reflect.Type]*refl.GroupedFieldsWithBS,
) *QueryResults {
    return &QueryResults{
        tx: tx,
        rows: rows,
        nextResult: nil,
        aliasedTables: aliasedTables,
        typeBSFieldMap: typeBSFieldMap,
        columns: nil,
        columnAliasFields: nil,
        closed: false,
    }
}
