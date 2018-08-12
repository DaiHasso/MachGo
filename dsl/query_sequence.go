package dsl

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"unicode"

	database "github.com/DaiHasso/MachGo"

	logging "github.com/daihasso/slogging"
	"github.com/jmoiron/sqlx"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var columnNamespaceRegex = regexp.MustCompile(`^([^\.]+)\.([^\.]+)$`)

// QuerySequence is an object that builds a query.
type QuerySequence struct {
	objects,
	joinedObjects []database.Object
	tableAliasMap  map[string]string
	aliasObjectMap map[string]database.Object
	columnAliasMap,
	aliasColumnMap map[string]string
	objectAliasCounter      int
	selectColumnExpressions []SelectColumnExpression
	whereBuilder            *WhereBuilder
	manager                 *database.Manager
}

// SelectColumnExpression is a select expression optionally tied to a table.
type SelectColumnExpression struct {
	isNamespaced bool
	columnName,
	tableNamespace string
}

func namespacedColumnFromString(raw string) NamespacedColumn {
	column := new(NamespacedColumn)
	column.isNamespaced = false
	column.columnName = raw
	if columnNamespaceRegex.MatchString(raw) {
		results := columnNamespaceRegex.FindStringSubmatch(
			raw,
		)
		column.isNamespaced = true
		column.tableNamespace = results[1]
		column.columnName = results[2]
	}

	return *column
}

type joinExpression struct {
	fromObject,
	toObject database.Object
	relationship *Relationship
}

func ObjectColumn(obj database.Object, column string) NamespacedColumn {
	columnString := column
	if unicode.IsUpper(rune(column[0])) {
		field, found := reflect.TypeOf(obj).FieldByName(column)
		if found {
			if tagValue, ok := field.Tag.Lookup(`db`); ok {
				columnString = tagValue
			}
		}
	}
	namespacedColumn := NamespacedColumn{
		true,
		columnString,
		"",
		obj.GetTableName(),
	}

	return namespacedColumn
}

func newQuerySequence() *QuerySequence {
	return &QuerySequence{
		make([]database.Object, 0),
		make([]database.Object, 0),
		make(map[string]string, 0),
		make(map[string]database.Object, 0),
		make(map[string]string, 0),
		make(map[string]string, 0),
		0,
		make([]SelectColumnExpression, 0),
		nil,
		nil,
	}
}

// NewJoin creates a new QuerySequence with a join.
func NewJoin(objects ...database.Object) *QuerySequence {

	querySequence := newQuerySequence()

	querySequence.Join(objects...)

	return querySequence
}

// Join adds object to the QuerySequence's join.
func (self *QuerySequence) Join(objects ...database.Object) *QuerySequence {
	self.addObjects(objects...)
	self.joinedObjects = append(self.joinedObjects, objects...)

	return self
}

func (self *QuerySequence) addObjects(objects ...database.Object) {
	self.makeAliasesForObjects(objects...)
	self.objects = append(
		self.objects,
		objects...,
	)
}

func (self *QuerySequence) makeAliasesForObjects(objects ...database.Object) {
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
func (self *QuerySequence) SelectObject(objects ...database.Object) *QuerySequence {
	columnExpressions := make([]SelectColumnExpression, 0)

	for _, object := range objects {
		columnExpression := new(SelectColumnExpression)
		columnExpression.isNamespaced = true
		columnExpression.columnName = "*"
		columnExpression.tableNamespace = object.GetTableName()

		columnExpressions = append(columnExpressions, *columnExpression)
	}

	self.selectColumnExpressions = columnExpressions

	return self
}

// Where adds a where clause from a WhereBuilder into the query.
func (self *QuerySequence) Where(wb *WhereBuilder) *QuerySequence {
	if self.whereBuilder != nil {
		logging.Warn("Overriding where clause in query.").Send()
	}

	self.whereBuilder = wb

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

func (self *QuerySequence) SetManager(manager *database.Manager) *QuerySequence {
	self.manager = manager

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

// TODO: Make a function for querying into a provided set of objects.

func (self QuerySequence) buildQuery() (string, []interface{}) {
	var selectString, fromString string
	if len(self.selectColumnExpressions) == 0 {
		allAliases := make([]string, 0)
		for alias, _ := range self.aliasObjectMap {
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

			if selectExp.isNamespaced {
				if alias, ok := self.tableAliasMap[selectExp.tableNamespace]; ok {
					line += fmt.Sprintf("%s.", alias)
				} else {
					line += fmt.Sprintf("%s.", selectExp.tableNamespace)
				}
			}

			line += selectExp.columnName
		}
		selectString += line
	}

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

	query := fmt.Sprintf(
		"SELECT %s FROM %s",
		selectString,
		fromString,
	)

	args := make([]interface{}, 0)
	if self.whereBuilder != nil {
		whereClause, whereArgs := self.whereBuilder.asQuery(&self)
		args = whereArgs
		query += " WHERE " + whereClause
	}

	return query, args
}

func (self *QuerySequence) solveJoin() []joinExpression {
	results := make([]joinExpression, 0)
	matches := make(map[database.Object][]database.Object, 0)
	for _, toObject := range self.joinedObjects {
		if _, ok := matches[toObject]; !ok {
			matches[toObject] = make([]database.Object, 0)
		} else {
			continue
		}
	ToLoop:
		for _, fromObject := range self.joinedObjects {
			if _, ok := matches[fromObject]; !ok {
				matches[fromObject] = make([]database.Object, 0)
			}

			if objectMatches, ok := matches[fromObject]; ok {
				for _, match := range objectMatches {
					if match == toObject {
						goto ToLoop
					}
				}
			}
			if objectMatches, ok := matches[toObject]; ok {
				for _, match := range objectMatches {
					if match == fromObject {
						goto ToLoop
					}
				}
			}
			if fromObject == toObject {
				continue
			}

			_, _, relationship, err := findRelationshipBetweenObjects(
				fromObject,
				toObject,
			)

			if err == nil {
				newJoinExp := joinExpression{
					fromObject,
					toObject,
					&relationship,
				}
				results = append(results, newJoinExp)
				matches[fromObject] = append(matches[fromObject], toObject)
				matches[toObject] = append(matches[toObject], fromObject)
				break
			}
		}
	}

	return results
}

func findRelationshipBetweenObjects(object1, object2 database.Object) (
	chosenObject,
	otherObject database.Object,
	chosenRelationship Relationship,
	err error,
) {
	isRelationshipable := false
	relationshipChosen := false

	if relationshipable, ok := object1.(Relationshipable); ok {
		isRelationshipable = true

		for _, relationship := range relationshipable.Relationships() {
			// TODO: Consider using reflected name to check for names as well.
			if relationship.Target == object2.GetTableName() {
				relationshipChosen = true
				chosenRelationship = relationship
				break
			}
		}
	}
	if relationshipable, ok := object2.(Relationshipable); ok {
		if !relationshipChosen {
			isRelationshipable = true

			for _, relationship := range relationshipable.Relationships() {
				// TODO: Consider using reflected name to check for names as well.
				if relationship.Target == object1.GetTableName() {
					relationshipChosen = true
					chosenRelationship = relationship
					break
				}
			}
		}
	}

	if !isRelationshipable {
		err = errors.New("none of the objects have relationships")
	} else if !relationshipChosen {
		err = errors.New(
			"no compatibile relationships for these two objects",
		)
	}

	return
}
