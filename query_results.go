package MachGo

import (
	"fmt"
	"reflect"

	"github.com/jmoiron/sqlx"

	"github.com/DaiHasso/MachGo/refl"
)

type QueryResults struct {
	tx *sqlx.Tx
	rows *sqlx.Rows
	nextResult *QueryResult
	aliasedObjects *AliasedObjects
	typeBSFieldMap map[reflect.Type]*refl.GroupedFieldsWithBS
	lastError error
	columns []string
	columnAliasFields []ColumnAliasField
}

func (self *QueryResults) Next() bool {
	hasRows := self.rows.Next()
	err := self.rows.Err()
	if !hasRows {
		if err != nil {
			self.lastError = err
		}
		return false
	}

	if self.columns == nil {
		columnNames, err := self.rows.Columns()
		if err != nil {
			self.lastError = err
			return false
		}

		self.columns = columnNames

		self.columnAliasFields, err = columnsToFieldNames(
			self.columns,
			self.typeBSFieldMap,
			self.aliasedObjects,
		)
		if err != nil {
			self.lastError = err
			return false
		}
	}

	self.nextResult, err = NewQueryResult(
		self.rows,
		self.aliasedObjects,
		self.columnAliasFields,
	)
	if err != nil {
		self.lastError = err
		return false
	}

	return true
}

func (self *QueryResults) GetResult() *QueryResult {
	return self.nextResult
}

func (self *QueryResults) Err() error {
	return self.lastError
}

func (self *QueryResults) WriteAllTo(objectSlices ...interface{}) error {
	elemTypes := make([]reflect.Type, len(objectSlices))
	for i, objectSlice := range objectSlices {
		objElemType, err := refl.ElementTypeFromSlice(objectSlice)
		if err != nil {
			return err
		}

		objElemTypeDeref := objElemType
		if objElemType.Kind() == reflect.Ptr {
			objElemTypeDeref = objElemType.Elem()
		}

		if _, ok := self.typeBSFieldMap[objElemTypeDeref]; !ok {
			return fmt.Errorf("No results match type '%s'.", objElemTypeDeref)
		}

		elemTypes[i] = objElemType
	}

	for self.Next() {
		elemValues := make([]interface{}, len(elemTypes))
		for i, elemTypePtr := range elemTypes {
			elemType := elemTypePtr.Elem()
			elemVal := reflect.New(elemType).Interface()
			elemValues[i] = elemVal
		}
		nextResult := self.GetResult()
		err := nextResult.WriteTo(elemValues...)
		if err != nil {
			return err
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

	return self.tx.Commit()
}

func NewQueryResults(
	tx *sqlx.Tx,
	rows *sqlx.Rows,
	objects []Object, // TODO: We should be receiving AliasedObjects instead.
	typeBSFieldMap map[reflect.Type]*refl.GroupedFieldsWithBS,
) (*QueryResults, error) {
	aliasedObjects, err := NewAliasedObjects(objects...)
	return &QueryResults{
		tx: tx,
		rows: rows,
		nextResult: nil,
		aliasedObjects: aliasedObjects,
		typeBSFieldMap: typeBSFieldMap,
		columns: nil,
		columnAliasFields: nil,
	}, err
}
