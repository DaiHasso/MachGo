package dsl

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	logging "github.com/daihasso/slogging"

	"MachGo"
	"MachGo/database"
	"MachGo/refl"
	"MachGo/pool"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var columnNamespaceRegex = regexp.MustCompile(`^([^\.]+)\.([^\.]+)$`)
var columnAliasNamespaceRegex = regexp.MustCompile(`^([^_]+)_(.*)$`)

// QuerySequence is an object that builds a query.
// TODO: All these maps should probably be something more clever.
// TODO: Refactor QuerySequence to use the new AliasedObjects struct to solve
//       above todo.
type QuerySequence struct {
	objects,
	joinedObjects []MachGo.Object
	typeAliasMap map[reflect.Type]string
	aliasTypeMap map[string]reflect.Type
	tableAliasMap map[string]string
	aliasObjectMap map[string]MachGo.Object
	columnAliasMap,
	aliasTableMap map[string]string
	typeBSFieldMap,
	typeFieldNameBSFieldMap map[reflect.Type]*refl.GroupedFieldsWithBS
	objectAliasCounter int
	selectColumnExpressions []SelectColumnExpression
	selectObjectTables map[string]int
	whereClause []Queryable
	manager *database.Manager
}

// SelectColumnExpression is a select expression optionally tied to a table.
type SelectColumnExpression struct {
	isNamespaced bool
	columnName,
	tableNamespace string
}

type joinExpression struct {
	fromObject,
	toObject MachGo.Object
	relationship *MachGo.Relationship
}

func newQuerySequence() *QuerySequence {
	return &QuerySequence{
		objects: make([]MachGo.Object, 0),
		joinedObjects: make([]MachGo.Object, 0),
		typeAliasMap: make(map[reflect.Type]string),
		aliasTypeMap: make(map[string]reflect.Type),
		tableAliasMap: make(map[string]string),
		aliasObjectMap: make(map[string]MachGo.Object),
		columnAliasMap: make(map[string]string),
		aliasTableMap: make(map[string]string),
		typeBSFieldMap: make(map[reflect.Type]*refl.GroupedFieldsWithBS),
		typeFieldNameBSFieldMap: make(
			map[reflect.Type]*refl.GroupedFieldsWithBS,
		),
		objectAliasCounter: 0,
		selectColumnExpressions: make([]SelectColumnExpression, 0),
		selectObjectTables: make(map[string]int),
		whereClause: nil,
		manager: nil,
	}
}

// NewQuerySequence creates a new empty QuerySequence.
func NewQuerySequence() *QuerySequence {

	querySequence := newQuerySequence()

	return querySequence
}

// NewJoin creates a new QuerySequence with a join.
func NewJoin(objects ...MachGo.Object) *QuerySequence {

	querySequence := newQuerySequence()

	manager, err := database.NewManager()
	if err != nil {
		// TODO: Maybe this is a panic? Or check for specific error.
		logging.Debug(
			"QuerySequence created without global manager set.",
		).Send()
	} else {
		querySequence.manager = manager
	}

	querySequence.Join(objects...)

	return querySequence
}

// Join adds object to the QuerySequence's join.
func (self *QuerySequence) Join(objects ...MachGo.Object) *QuerySequence {
	self.addObjects(objects...)
	self.joinedObjects = append(self.joinedObjects, objects...)

	return self
}

func (self *QuerySequence) addObjects(objects ...MachGo.Object) {
	self.makeAliasesForObjects(objects...)
	self.objects = append(
		self.objects,
		objects...,
	)
	self.cacheTagsForType(objects...)
}

func (self *QuerySequence) cacheTagsForType(objects ...MachGo.Object) {
	for _, object := range objects {
		objType := refl.Deref(reflect.TypeOf(object))
		fieldGroupings := refl.GetGroupedFieldsWithBS(
			object,
			refl.GroupFieldsByTagValue("db", "dbfkey"),
			refl.GroupFieldsByFieldName(),
		)
		tagValBSFields := fieldGroupings[0]
		fieldNameBSFields := fieldGroupings[1]

		self.typeBSFieldMap[objType] = tagValBSFields
		self.typeFieldNameBSFieldMap[objType] = fieldNameBSFields
	}
}

func (self *QuerySequence) makeAliasesForObjects(objects ...MachGo.Object) {
	for _, object := range objects {
		if self.objectAliasCounter > len(alphabet) {
			// TODO: Make this account for doubled aliases like aa, ab, etc.
			panic(
				"There's more objects in this query than we've accounted for.",
			)
		}

		alias := string(alphabet[self.objectAliasCounter])
		self.aliasObjectMap[alias] = object
		self.tableAliasMap[object.GetTableName()] = alias
		self.aliasTableMap[alias] = object.GetTableName()

		// NOTE: Does the type map cover all the uses for the object map?
		// TODO: Check if ptr before indirecting.
		objType := reflect.Indirect(reflect.ValueOf(object)).Type()
		self.typeAliasMap[objType] = alias
		self.aliasTypeMap[alias] = objType

		self.objectAliasCounter++
	}
}

// Select sets the columns to select for a QuerySequence.
func (self *QuerySequence) Select(columns ...string) *QuerySequence {
	columnExpressions := make([]SelectColumnExpression, 0)

	for _, column := range columns {
		columnExpression := new(SelectColumnExpression)
		columnExpression.isNamespaced = false
		columnExpression.columnName = column
		columnExpression.tableNamespace = ""

		if columnNamespaceRegex.MatchString(column) {
			results := columnNamespaceRegex.FindStringSubmatch(
				column,
			)
			columnExpression.isNamespaced = true
			columnExpression.tableNamespace = results[1]
			columnExpression.columnName = results[2]
		}

		columnExpressions = append(columnExpressions, *columnExpression)
	}

	self.selectColumnExpressions = columnExpressions

	return self
}

// SelectObject selects data pertaining to the given object.
func (self *QuerySequence) SelectObject(
	objects ...MachGo.Object,
) *QuerySequence {
	columnExpressions := make([]SelectColumnExpression, 0)

	for i, object := range objects {
		columnExpression := new(SelectColumnExpression)
		columnExpression.isNamespaced = true
		columnExpression.columnName = "*"
		columnExpression.tableNamespace = object.GetTableName()

		columnExpressions = append(columnExpressions, *columnExpression)

		self.selectObjectTables[columnExpression.tableNamespace] = i
	}

	self.selectColumnExpressions = columnExpressions


	return self
}

// Where adds a where clause to the query.
func (self *QuerySequence) Where(
	firstQuery Queryable, rest ...Queryable,
) *QuerySequence {
	clauses := append([]Queryable{firstQuery}, rest...)
	self.whereClause = append(self.whereClause, clauses...)

	return self
}

// PrintQuery print the evaluated query.
func (self QuerySequence) PrintQuery() string {
	query, args := self.buildQuery()
	queryString := fmt.Sprintf(
		`query: "%s", args: (%+v)`,
		query,
		args,
	)

	return queryString
}

func (self *QuerySequence) SetManager(
	manager *database.Manager,
) *QuerySequence {
	self.manager = manager

	return self
}

func (self *QuerySequence) SetPool(
	connPool *pool.ConnectionPool,
) *QuerySequence {
	var err error
	self.manager, err = database.NewManagerFromPool(connPool)
	if err != nil {
		panic(err)
	}

	return self
}

// Query uses a transaction to get and return rows, caller is expected to
// manager the transaction.
func (self QuerySequence) Query(tx *sqlx.Tx) (*sqlx.Rows, error) {
	var err error
	var rows *sqlx.Rows
	query, variables := self.buildQuery()

	query = self.manager.Rebind(query)

	rows, err = tx.Queryx(query, variables...)
	if err != nil {
		return nil, err
	}

	return rows, nil
}

// QueryInterface gets rows from the database and translates them to generic
// slices of slices of interfaces.
func (self QuerySequence) QueryInterface() ([][]interface{}, error) {
	results := make([][]interface{}, 0)

	action := func(tx *sqlx.Tx) error {
		rows, err := self.Query(tx)
		if err != nil {
			return err
		}

		for rows.Next() {
			cols, err := rows.SliceScan()
			if err != nil {
				return err
			}
			results = append(results, cols)
		}

		return nil
	}

	return results, self.manager.Transactionized(action)
}

// IntoObjects reads the query results and produces objects if it finds matches
// in its known objects.
// TODO: This probably should only check the selected objects not all known
//       objects.
// TODO: This should complain about or handle mixes of non-namespaced and
//       namespaced content.
// TODO: This needs optimization to not be checking all the columns against all
//       the known objects for every row.
// TODO: This needs to signify in some way what order the returned objects are
//       in or maybe just return them in select-order.
func (self QuerySequence) IntoObjects() ([][]interface{}, error) {
	if self.manager == nil {
		return nil, errors.New("Manager hasn't been set.")
	}

	var objs [][]interface{}

	action := func(tx *sqlx.Tx) error {
		query, variables := self.buildQuery()

		query = self.manager.Rebind(query)

		rows, err := tx.Query(query, variables...)
		if err != nil {
			return err
		}

		hasRows := rows.Next()
		if ! hasRows {
			return nil
		}

		columnNames, err := rows.Columns()
		if err != nil {
			return err
		}

		resultTypes := make(map[string]reflect.Type)
		aliasIndexes := make(map[string]int)
		for _, column := range columnNames {
			columnIdent := ColumnIdentifierFromResult(column)
			if tableAlias, ok := columnIdent.Scope(); ok {
				namespaceType := self.aliasTypeMap[tableAlias]
				if namespaceType == nil {
					return fmt.Errorf(
						"An object with the alias '%s' hasn't been added to "+
							"the query.",
						tableAlias,
					)
				}

				tableName := self.aliasTableMap[tableAlias]
				index, inSelect := self.selectObjectTables[tableName]
				if !inSelect {
					return errors.New("Data not in select returned.")
				}

				if _, seen := aliasIndexes[tableAlias]; !seen {
					resultTypes[tableAlias] = namespaceType
					aliasIndexes[tableAlias] = index
				}
			}
		}

		if len(aliasIndexes) < len(self.selectObjectTables) {
			return fmt.Errorf(
				"Returned data had %d tables in returned data but expected "+
					"%d tables.",
				len(aliasIndexes),
				len(self.selectObjectTables),
			)
		}

		for hasRows {
			values := make([]interface{}, 0)
			rowObjs := make([]interface{}, len(aliasIndexes))
			aliasVals := make(map[string]reflect.Value, len(aliasIndexes))
			for alias, objType := range resultTypes {
				aliasVals[alias] = reflect.New(objType)
			}

			for _, column := range columnNames {
				columnIdent := ColumnIdentifierFromResult(column)
				if tableAlias, ok := columnIdent.Scope(); ok {
					index := aliasIndexes[tableAlias]

					objVal := aliasVals[tableAlias]
					objPtr := objVal.Interface()
					objCoerced := objPtr
					rowObjs[index] = objCoerced

					objType := refl.Deref(objVal.Type())
					tagValBSFields := *self.typeBSFieldMap[objType]

					var fieldName string
					columnName := columnIdent.Column()
					bsField, fieldExists := tagValBSFields[columnName]
					if columnName == "id" {
						// TODO: Fix this hacky usecase. It has something to do
						//       with the nested struct not populating tags
						//       maybe?
						fieldName = strings.ToUpper(columnName)
					} else if fieldExists{
						fieldName = bsField.Name()
					} else {
						fieldName = refl.LowerSnakeToUpperCamel(columnName)
					}

					field := reflect.Indirect(objVal).FieldByName(fieldName)

					if field.IsValid() {
						// Stole from reflectx: https://tinyurl.com/yc3lpeam
						if field.Kind() == reflect.Ptr && field.IsNil() {
							alloc := reflect.New(refl.Deref(field.Type()))
							field.Set(alloc)
						}
						if field.Kind() == reflect.Map && field.IsNil() {
							field.Set(reflect.MakeMap(field.Type()))
						}

						values = append(values, field.Addr().Interface())
					} else {
						int := new(interface{})
						values = append(values, int)
					}
				} else {
					// Throw it into the ether.
					int := new(interface{})
					values = append(values, int)
				}
			}

			err = rows.Scan(values...)
			if err != nil {
				return err
			}

			objs = append(objs, rowObjs)

			hasRows = rows.Next()
		}

		return nil
	}

	return objs, self.manager.Transactionized(action)
}

func (self QuerySequence) Results() (*MachGo.QueryResults, error) {
	if self.manager == nil {
		return nil, errors.New("Manager hasn't been set.")
	}

	tx, err := self.manager.Beginx()
	if err != nil {
		return nil, err
	}

	query, variables := self.buildQuery()

	query = self.manager.Rebind(query)

	rows, err := tx.Queryx(query, variables...)
	if err != nil {
		newErr := tx.Rollback()
		if newErr != nil {
			return nil, newErr
		}

		return nil, err
	}

	aliasedObjects, err := MachGo.NewAliasedObjectsFromExisting(
		self.aliasObjectMap,
	)
	if err != nil {
		return nil, err
	}

	return MachGo.NewQueryResults(
		tx,
		rows,
		aliasedObjects,
		self.typeBSFieldMap,
	), nil
}

// NOTE: I think reflectx has something like this, should we be using it?
func (self QuerySequence) getSelectableColumns(obj MachGo.Object) []string {
	objType := refl.Deref(reflect.TypeOf(obj))
	bsFieldMap := self.typeFieldNameBSFieldMap[objType]

	objAlias := self.tableAliasMap[obj.GetTableName()]
	var columns []string
	for _, bsField := range *bsFieldMap {
		var foreignAlias string
		var column string
		if bsTag := bsField.Tag("db"); bsTag != nil {
			column = bsTag.Value()
			foreignAlias = objAlias
			if bsTag.HasProperty("foreign") {
				foreignBSTag := bsField.Tag("dbforeign")
				if foreignBSTag == nil {
					panic(fmt.Errorf(
						"Foreign column declared without 'dbforeign' tag " +
							"declared for field '%s'",
						bsField.Name(),
					))
				}

				foreignTable := foreignBSTag.Value()
				var ok bool
				foreignAlias, ok = self.tableAliasMap[foreignTable]
				if !ok {
					panic(fmt.Errorf(
						"Table '%s' declared in foreign relationship is not " +
						"in QuerySequence.",
						foreignTable,
					))
				}
			}
			aliasedColumn := fmt.Sprintf(
				"%s.%s as %s_%s",
				foreignAlias,
				column,
				objAlias,
				column,
			)
			columns = append(columns, aliasedColumn)
		}
	}

	// TODO: This is hacky, figure out something more elegant.
	// NOTE: Maybe they could be optional function calls on an Object? Some
	//       kind of "default_attr" interface?
	if _, ok := obj.(MachGo.IDAttribute); ok {
		newColumn := fmt.Sprintf("%s.id as %s_id", objAlias, objAlias)
		columns = append(columns, newColumn)
	}

	if _, ok := obj.(MachGo.DefaultAttributes); ok {
		createdColumn := fmt.Sprintf(
			"%s.created as %s_created",
			objAlias,
			objAlias,
		)
		updatedColumn := fmt.Sprintf(
			"%s.updated %s_updated",
			objAlias,
			objAlias,
		)
		columns = append(columns, createdColumn)
		columns = append(columns, updatedColumn)
	}

	return columns
}

// TODO: Make a function for querying into a provided set of objects.

func (self QuerySequence) buildQuery() (string, []interface{}) {
	var selectString, fromString string
	if len(self.selectColumnExpressions) == 0 {
		allAliases := make([]string, 0)
		for alias := range self.aliasObjectMap {
			allAliases = append(allAliases, alias)
		}
		// Sorting this should make this more easily testable.
		// TODO: Assess performance.
		sort.Strings(allAliases)
		for _, alias := range allAliases {
			if len(selectString) > 0 {
				selectString += ", "
			}
			selectString += fmt.Sprintf("%s.*", alias)
		}
	} else {
		line := ""
		for _, selectExp := range self.selectColumnExpressions {
			if len(line) != 0 {
				line += ", "
			}

			if selectExp.columnName == "*" {
				alias := self.tableAliasMap[selectExp.tableNamespace]
				obj := self.aliasObjectMap[alias]
				line += strings.Join(self.getSelectableColumns(obj), ", ")
			} else {
				if selectExp.isNamespaced {
					alias, ok := self.tableAliasMap[selectExp.tableNamespace]
					if ok {
						line += fmt.Sprintf("%s.", alias)
					} else {
						line += fmt.Sprintf("%s.", selectExp.tableNamespace)
					}
				}

				line += selectExp.columnName
			}
		}
		selectString += line
	}

	if len(self.joinedObjects) == 1 {
		// It's just a normal select with no join.
		onlyObject := self.joinedObjects[0]
		objectAlias := self.tableAliasMap[onlyObject.GetTableName()]
		fromString = fmt.Sprintf(
			"%s %s",
			onlyObject.GetTableName(),
			objectAlias,
		)
	} else {
		joinExps := self.solveJoin()
		for _, exp := range joinExps {
			fromAlias := self.tableAliasMap[exp.fromObject.GetTableName()]
			toAlias := self.tableAliasMap[exp.toObject.GetTableName()]
			line := ""
			if len(fromString) == 0 {
				line += fmt.Sprintf(
					"%s %s",
					exp.fromObject.GetTableName(),
					fromAlias,
				)
			}
			line += fmt.Sprintf(
				" JOIN %s %s ON %s.%s=%s.%s",
				exp.toObject.GetTableName(),
				toAlias,
				fromAlias,
				exp.relationship.SelfColumn,
				toAlias,
				exp.relationship.TargetColumn,
			)
			fromString += line
		}
	}

	// #nosec G201
	query := fmt.Sprintf(
		"SELECT %s FROM %s",
		selectString,
		fromString,
	)

	args := make([]interface{}, 0)
	if self.whereClause != nil {
		whereTemplate := `WHERE %s`
		combinedWhere := NewMultiAndCondition(self.whereClause...)
		whereClause, whereArgs := combinedWhere.QueryValue(&self)
		args = whereArgs
		query += " " + fmt.Sprintf(whereTemplate, whereClause)
	}

	return query, args
}

func (self *QuerySequence) solveJoin() []joinExpression {
	results := make([]joinExpression, 0)
	matches := make(map[string][]MachGo.Object)
	for _, toObject := range self.joinedObjects {
		if _, ok := matches[toObject.GetTableName()]; !ok {
			matches[toObject.GetTableName()] = make([]MachGo.Object, 0)
		} else {
			continue
		}
	ToLoop:
		for _, fromObject := range self.joinedObjects {
			if _, ok := matches[fromObject.GetTableName()]; !ok {
				matches[fromObject.GetTableName()] = make([]MachGo.Object, 0)
			}

			if objectMatches, ok := matches[fromObject.GetTableName()]; ok {
				for _, match := range objectMatches {
					if match == toObject {
						goto ToLoop
					}
				}
			}
			if objectMatches, ok := matches[toObject.GetTableName()]; ok {
				for _, match := range objectMatches {
					if match == fromObject {
						goto ToLoop
					}
				}
			}
			if fromObject == toObject {
				continue
			}

			joinExp, err := findRelationshipBetweenObjects(
				fromObject,
				toObject,
			)

			if err == nil {
				results = append(results, *joinExp)
				matches[fromObject.GetTableName()] = append(
					matches[fromObject.GetTableName()],
					toObject,
				)
				matches[toObject.GetTableName()] = append(
					matches[toObject.GetTableName()],
					fromObject,
				)
				break
			}
		}
	}

	return results
}

// TODO: Audit performance. Consider short-circut conditions.
func findRelationshipBetweenObjects(object1, object2 MachGo.Object) (
	*joinExpression,
	error,
) {
	isRelationshipable := false
	matches := make(map[MachGo.Object][]*joinExpression)

	if relationshipable, ok := object1.(MachGo.Relationshipable); ok {
		isRelationshipable = true

		for _, relationship := range relationshipable.Relationships() {
			// TODO: Consider using reflected name to check for names as well.
			if relationship.TargetTable() == object2.GetTableName() {
				joinExp := &joinExpression{
					object1,
					object2,
					&relationship,
				}
				if matchExps, ok := matches[object1]; ok {
					matches[object1] = append(matchExps, joinExp)
				} else {
					matchExps := make([]*joinExpression, 1)
					matchExps[0] = joinExp
					matches[object1] = matchExps
				}
			}
		}
	}
	if relationshipable, ok := object2.(MachGo.Relationshipable); ok {
		for _, relationship := range relationshipable.Relationships() {
			// TODO: Consider using reflected name to check for names as well.
			if relationship.TargetTable() == object1.GetTableName() {
				joinExp := &joinExpression{
					object2,
					object1,
					&relationship,
				}
				if matchExps, ok := matches[object2]; ok {
					matches[object2] = append(matchExps, joinExp)
				} else {
					matchExps := make([]*joinExpression, 1)
					matchExps[0] = joinExp
					matches[object2] = matchExps
				}
			}
		}
	}

	if joinExps, ok := matches[object1]; ok {
		return joinExps[0], nil
	} else if joinExps, ok := matches[object2]; ok {
		joinExp := joinExps[0]
		newJoinExp := &joinExpression{
			toObject: joinExp.fromObject,
			fromObject: joinExp.toObject,
			relationship: joinExp.relationship.Invert(),
		}
		return newJoinExp, nil
	}

	if !isRelationshipable {
		return nil, errors.New("none of the objects have relationships")
	}

	return nil, errors.New(
		"no compatibile relationships for these two objects",
	)
}

func (self *QuerySequence) AliasForTable(tableName string) string {
	if alias, ok := self.tableAliasMap[tableName]; ok {
		return alias
	}

	return tableName
}
