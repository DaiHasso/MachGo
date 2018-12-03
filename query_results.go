package MachGo

import (
	"fmt"
	"reflect"
	"strings"

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
	aliasesInSelect map[string]bool
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

		columnAliasFields, aliasesInSelect, err := columnsToFieldNames(
			self.columns,
			self.typeBSFieldMap,
			self.aliasedObjects,
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

func (self *QueryResults) WriteAllTo(
	objectSlices ...interface{},
) (retErr error) {
	defer func() {
		retErr = self.tx.Commit()
	}()

	elemTypes := make([]reflect.Type, len(objectSlices))
	for i, objectSlice := range objectSlices {
		if typ := reflect.TypeOf(objectSlice); typ.Kind() != reflect.Ptr {
			return fmt.Errorf(
				"Type must be pointer to slice not %T.",
				objectSlice,
			)
		}
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
			elemAlias, err := self.aliasedObjects.ObjectAlias(
				elemVal.(Object),
			)
			if err != nil {
				return err
			}
			if _, ok := self.aliasesInSelect[elemAlias]; !ok {
				return fmt.Errorf(
					"Can't write slice of type '%s', object is not in select.",
					elemType,
				)
			}

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

	return nil
}

func columnsToFieldNames(
    columnNames []string,
    typeBSFieldMap map[reflect.Type]*refl.GroupedFieldsWithBS,
    aliasedObjects *AliasedObjects,
) ([]ColumnAliasField, map[string]bool, error) {
    columnAliasFields := make([]ColumnAliasField, len(columnNames))
    aliasesInSelect := make(map[string]bool, len(columnNames))
    for i, column := range columnNames {
        columnAlias, ok := ColumnAliasFromString(column)
        if !ok {
            return nil, nil, fmt.Errorf(
				"Unexpected column in result: '%s'",
				column,
			)
        }

		aliasesInSelect[columnAlias.TableAlias] = true

        objType := aliasedObjects.TypeForAlias(columnAlias.TableAlias)
        tagValBSFields := *typeBSFieldMap[*objType]

        var fieldName string
        if columnAlias.ColumnName == "id" {
            // TODO: Fix this hacky usecase. It has something to do
            //       with the nested struct not populating tags
            //       maybe?
            fieldName = strings.ToUpper(columnAlias.ColumnName)
        } else if bsField, ok := tagValBSFields[columnAlias.ColumnName]; ok {
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


func NewQueryResults(
	tx *sqlx.Tx,
	rows *sqlx.Rows,
	aliasedObjects *AliasedObjects,
	typeBSFieldMap map[reflect.Type]*refl.GroupedFieldsWithBS,
) *QueryResults {
	return &QueryResults{
		tx: tx,
		rows: rows,
		nextResult: nil,
		aliasedObjects: aliasedObjects,
		typeBSFieldMap: typeBSFieldMap,
		columns: nil,
		columnAliasFields: nil,
	}
}
