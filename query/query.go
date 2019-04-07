package query

import (
    "database/sql"
    "fmt"
    "reflect"
    "sort"
    "strings"
   
    "github.com/pkg/errors"

    "github.com/daihasso/machgo/refl"
    "github.com/daihasso/machgo/query/qtypes"
    "github.com/daihasso/machgo/pool"
    "github.com/daihasso/machgo/base"
)

type queryValuePair struct {
    query string
    values []interface{}
    valid bool
}

func (self *queryValuePair) invalidate() {
    self.query = ""
    self.values = nil
    self.valid = false
}

type cachedQuery struct {
    Select queryValuePair
    From queryValuePair
    Where queryValuePair
    Options queryValuePair
}

// Query represents an in-flight query that is being built as before actually
// making the request to the DB. It's purpose is to facillitate a more
// straightforward interaction with the DB for general purpose usage to avoid
// having to write direct SQL.
type Query struct {
    Pool *pool.ConnectionPool
    Tables *qtypes.AliasedTables
    WhereClauses []qtypes.Queryable
    SelectExpressions []qtypes.SelectExpression
    OptionClauses []qtypes.QueryOption

    typeBSFieldMap,
    typeFieldNameBSFieldMap map[reflect.Type]*refl.GroupedFieldsWithBS
    joinedObjects []base.Base

    cached cachedQuery

    Errors []error
}

func (self *Query) addObjects(objects []base.Base) error {
    err := self.Tables.AddObjects(objects...)
    if err != nil {
        return errors.Wrap(
            err, "Error while adding objects to query",
        )
    }
    self.joinedObjects = append(self.joinedObjects, objects...)
    self.cacheTagsForType(objects)

    return nil
}

func (self *Query) cacheTagsForType(objects []base.Base) {
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

// Join adds an object to the query. The joined object should have a
// Relationship with at least one other object already in the Query.
func (self *Query) Join(objects ...base.Base) *Query {
    self.cached.Select.invalidate()
    self.cached.From.invalidate()

    err := self.addObjects(objects)
    if err != nil {
        wrapped := errors.Wrap(err, "Error while adding objects to join")
        self.Errors = append(self.Errors, wrapped)
    }

    return self
}

// Where creates or appends to the where clause the provided clauses.
func (self *Query) Where(clauses ...qtypes.Queryable) *Query {
    self.cached.Where.invalidate()

    self.WhereClauses = append(self.WhereClauses, clauses...)

    return self
}

// Select sets the selected data for the Query. Repeated calls to this
// function will Override the existing select clause.
func (self *Query) Select(stmnts ...qtypes.Selectable) *Query {
    self.cached.Select.invalidate()

    // NOTE: We nuke existing select expression effectively overwriting each
    //       time this is called.
    var selectExpressions []qtypes.SelectExpression
    for _, statement := range stmnts {
        stmnt, err := statement()
        if err == nil {
            selectExpressions = append(selectExpressions, stmnt)
        } else {
            wrapped := errors.Wrap(err, "Error evaluating select statement")
            self.Errors = append(self.Errors, wrapped)
        }
    }
    self.SelectExpressions = selectExpressions

    return self
}

// Limit sets a limit to the total returned items for the query.
func (self *Query) Limit(limit int) *Query {
    self.cached.Options.invalidate()

    limitOption := qtypes.LimitOption{
        Limit: limit,
    }
    for i, optionClause := range self.OptionClauses {
        if optionClause.OptionType() == qtypes.LimitOptionType {
            self.OptionClauses[i] = limitOption
            return self
        }
    }

    self.OptionClauses = append(self.OptionClauses, limitOption)

    return self
}

// Offset sets an offset from the first item returned in the query.
func (self *Query) Offset(offset int) *Query {
    self.cached.Options.invalidate()

    offsetOption := qtypes.OffsetOption{
        Offset: offset,
    }
    for i, optionClause := range self.OptionClauses {
        if optionClause.OptionType() == qtypes.OffsetOptionType {
            self.OptionClauses[i] = offsetOption
            return self
        }
    }

    self.OptionClauses = append(self.OptionClauses, offsetOption)

    return self
}

// OrderBy adds an ordering to the query. Repeated calls to this function will
// append to the existing ordering.
func (self *Query) OrderBy(orderStatements ...qtypes.Queryable) *Query {
    self.cached.Options.invalidate()

    for i, optionClause := range self.OptionClauses {
        if optionClause.OptionType() == qtypes.OrderByOptionType {
            // If it's not the first call to order by add this order as a
            // new statement in the order clause. This behaviour is unique
            // to this option.
            orderOption := (self.OptionClauses[i]).(qtypes.OrderByOption)
            orderOption.AddStatements(orderStatements...)
            self.OptionClauses[i] = orderOption
            return self
        }
    }

    orderOption := qtypes.OrderByOption{
        Order: qtypes.NewMultiListCondition(orderStatements...),
    }

    self.OptionClauses = append(self.OptionClauses, orderOption)

    return self
}

// Order is a convenience wrapper for OrderBy.
func (self *Query) Order(order ...qtypes.Queryable) *Query {
    return self.OrderBy(order...)
}

// String wraps the PrintQuery method so that auto-stringing prints the
// resulting query itself instead of the object.
func (self Query) String() string {
    return self.PrintQuery()
}

// PrintQuery prints what query will be made when the query is called.
func (self Query) PrintQuery() string {
    if len(self.Errors) != 0 {
        return fmt.Sprintf(
            "Error building query: '%#+v'",
            self.Errors,
        )
    }

    query, args, err := self.buildQuery()
    if err != nil {
        return fmt.Sprintf(
            "Error building query: '%s'",
            err.Error(),
        )
    }
    var argsStrings []string
    for _, arg := range args {
        var argString string
        if namedArg, ok := arg.(sql.NamedArg); ok {
            argString = fmt.Sprintf(
                "%s: %#+v", namedArg.Name, namedArg.Value,
            )
        } else if stringer, ok := arg.(fmt.Stringer); ok {
            argString = stringer.String()
        } else {
            argString = fmt.Sprintf("%#+v", arg)
        }

        argsStrings = append(argsStrings, argString)
    }
    argsString := strings.Join(argsStrings, ", ")
    queryString := fmt.Sprintf(
        `query: '%s', args: (%s)`, query, argsString,
    )

    return queryString
}

// Results returns the query as a QueryResults object that can be used to read
// data out of the database in a convenient fashion.
func (self Query) Results() (*qtypes.QueryResults, error) {
    if len(self.Errors) != 0 {
        return nil, errors.Errorf(
            "Errors while forming query:\n%#+v",
            self.Errors,
        )
    }

    tx, err := self.Pool.Beginx()
    if err != nil {
        return nil, err
    }

    query, variables, err := self.buildQuery()
    if err != nil {
        return nil, errors.Wrap(
            err, "Error while building query",
        )
    }

    query = self.Pool.Rebind(query)

    variableMap := make(map[string]interface{}, len(variables))
    for _, variable := range variables {
        if namedVar, ok := variable.(sql.NamedArg); ok {
            variableMap[namedVar.Name] = namedVar.Value
        }
    }

    rows, err := tx.NamedQuery(query, variableMap)
    if err != nil {
        newErr := tx.Rollback()
        if newErr != nil {
            return nil, newErr
        }

        return nil, err
    }

    return qtypes.NewQueryResults(
        tx, rows, self.Tables, self.typeBSFieldMap,
    ), nil
}

// Count makes a call to the db to get the total count that would be returned
// from the query without any extra QueryOptions (limit, offset, etc).
func (self Query) Count() (count int, err error) {
    if len(self.Errors) != 0 {
        return -1, errors.Errorf(
            "Errors while forming query:\n%#+v",
            self.Errors,
        )
    }

    starLiteral := qtypes.LiteralQueryable{
        Value: "*",
    }
    selectFunc := qtypes.SelectCount{
        Expression: starLiteral,
    }
    selectString, args := selectFunc.QueryValue(self.Tables)

    fromString, err := self.buildFrom()
    if err != nil {
        return -1, errors.Wrap(err, "Error while building from clause")
    }

    // #nosec G201
    query := fmt.Sprintf(
        "SELECT %s FROM %s",
        selectString,
        fromString,
    )

    whereQuery, whereArgs := self.buildWhere()
    if whereQuery != "" {
        query += " " + whereQuery
        args = append(args, whereArgs...)
    }


    tx, err := self.Pool.Beginx()
    if err != nil {
        return -1, err
    }

    query = self.Pool.Rebind(query)

    variableMap := make(map[string]interface{}, len(args))
    for _, variable := range args {
        if namedVar, ok := variable.(sql.NamedArg); ok {
            variableMap[namedVar.Name] = namedVar.Value
        }
    }

    rows, err := tx.NamedQuery(query, variableMap)
    if err != nil {
        newErr := tx.Rollback()
        if newErr != nil {
            return -1, newErr
        }

        return -1, err
    }

    defer func() {
        rows.Close()
        if err != nil {
            newErr := tx.Rollback()
            if newErr != nil {
                err = newErr
            }
            return
        }

        err = tx.Commit()
        if err != nil {
            newErr := tx.Rollback()
            if newErr != nil {
                err = newErr
            }
        }
    }()

    if rows.Next() {
        err = rows.Scan(&count)
    }

    return count, nil
}


func (self Query) buildSelect() (string, error) {
    if self.cached.Select.valid {
        return self.cached.Select.query, nil
    }

    var selectString string
    if len(self.SelectExpressions) == 0 {
        allAliases := make([]string, 0)
        for _, alias := range self.Tables.Aliases() {
            allAliases = append(allAliases, alias)
        }
        // Sorting this should make this more easily testable.
        // TODO: Assess performance.
        sort.Strings(allAliases)
        for _, alias := range allAliases {
            if len(selectString) > 0 {
                selectString += ", "
            }
            typ := self.Tables.TypeForAlias(alias)
            columns, err := self.getSelectableColumns(*typ)
            if err != nil {
                return "", errors.Wrap(
                    err, "Error while trying to get columns for object",
                )
            }
            sort.Strings(columns)
            selectString += strings.Join(columns, ", ")
        }
    } else {
        line := ""
        for _, selectExp := range self.SelectExpressions {
            if len(line) != 0 {
                line += ", "
            }

            if selectExp.Column() == "*" {
                tableName, _ := selectExp.Table()
                typ := self.Tables.TypeForTable(tableName)
                columns, err := self.getSelectableColumns(*typ)
                if err != nil {
                    return "", errors.Wrap(
                        err, "Error while trying to get columns for object",
                    )
                }
                sort.Strings(columns)
                line += strings.Join(columns, ", ")
            } else {
                if tableName, ok := selectExp.Table(); ok {
                    alias, ok := self.Tables.AliasForTable(tableName)
                    if ok {
                        line += fmt.Sprintf("%s.", alias)
                    } else {
                        line += fmt.Sprintf("%s.", tableName)
                    }
                }

                line += selectExp.Column()
            }
        }
        selectString += line
    }

    self.cached.Select.query = selectString
    self.cached.Select.valid = true

    return selectString, nil
}

func (self Query) buildFrom() (string, error) {
    if self.cached.From.valid {
        return self.cached.From.query, nil
    }

    var fromString string
    if len(self.joinedObjects) == 1 {
        // It's just a normal select with no join.
        onlyObject := self.joinedObjects[0]
        tableName, err := base.BaseTable(onlyObject)
        if err != nil {
            return "", errors.Wrap(
                err, "Error while getting object table name",
            )
        }
        objectAlias, _ := self.Tables.AliasForTable(tableName)
        fromString = fmt.Sprintf("%s %s", tableName, objectAlias)
    } else {
        joinRels := self.solveJoin()
        if len(joinRels) == 0 {
            return "", errors.New(
                "Objects joined don't have relationships with each other",
            )
        }
        for _, rel := range joinRels {
            fromTable, toTable := rel.Tables()
            fromColumn, toColumn := rel.Columns()
            fromAlias, _ := self.Tables.AliasForTable(fromTable)
            toAlias, _ := self.Tables.AliasForTable(toTable)
            line := ""
            if len(fromString) == 0 {
                line += fmt.Sprintf(
                    "%s %s",
                    fromTable,
                    fromAlias,
                )
            }
            line += fmt.Sprintf(
                " JOIN %s %s ON %s.%s=%s.%s",
                toTable,
                toAlias,
                fromAlias,
                fromColumn,
                toAlias,
                toColumn,
            )
            fromString += line
        }
    }

    self.cached.From.query = fromString
    self.cached.From.valid = true

    return fromString, nil
}

func (self Query) buildWhere() (string, []interface{}) {
    if self.cached.Where.valid {
        return self.cached.Where.query, self.cached.Where.values
    }

    if len(self.WhereClauses) != 0 {
        whereTemplate := `WHERE %s`
        combinedWhere := qtypes.NewMultiAndCondition(self.WhereClauses...)
        whereClause, whereArgs := combinedWhere.QueryValue(self.Tables)
        whereString := fmt.Sprintf(whereTemplate, whereClause)

        self.cached.Where.query = whereString
        self.cached.Where.values = whereArgs
        self.cached.Where.valid = true
        return whereString, whereArgs
    }

    return "", nil
}

func (self Query) buildOptions() (string, []interface{}) {
    if self.cached.Options.valid {
        return self.cached.Options.query, self.cached.Options.values
    }

    var (
        qStrings []string
        args []interface{}
    )
    if len(self.OptionClauses) != 0 {
        sort.SliceStable(self.OptionClauses, func(i int, j int) bool {
            optTypeI := self.OptionClauses[i].OptionType()
            optTypeJ := self.OptionClauses[j].OptionType()
            return optTypeI < optTypeJ
        })

        for _, option := range(self.OptionClauses) {
            qString, qVal := option.QueryValue(self.Tables)
            qStrings = append(qStrings, qString)
            args = append(args, qVal...)
        }
    }

    optionQuery := strings.Join(qStrings, " ")

    self.cached.Options.query = optionQuery
    self.cached.Options.values = args
    self.cached.Options.valid = true

    return optionQuery, args
}

func (self Query) buildQuery() (string, []interface{}, error) {
    selectString, err := self.buildSelect()
    if err != nil {
        return "", nil, errors.Wrap(err, "Error while building select clause")
    }

    fromString, err := self.buildFrom()
    if err != nil {
        return "", nil, errors.Wrap(err, "Error while building from clause")
    }

    // #nosec G201
    query := fmt.Sprintf(
        "SELECT %s FROM %s",
        selectString,
        fromString,
    )

    whereQuery, args := self.buildWhere()
    if whereQuery != "" {
        query += " " + whereQuery
    }

    optionQuery, optArgs := self.buildOptions()
    if optionQuery != "" {
        query += " " + optionQuery
        args = append(args, optArgs...)
    }

    return query, args, nil
}

func (self Query) getSelectableColumns(typ reflect.Type) ([]string, error) {
    bsFieldMap := self.typeFieldNameBSFieldMap[typ]
    objTable := self.Tables.TypeTable(typ)
    objAlias, ok := self.Tables.AliasForTable(objTable)
    if !ok {
        objAlias = objTable
    }

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
                    panic(errors.Errorf(
                        "Foreign column declared without 'dbforeign' tag " +
                            "declared for field '%s'",
                        bsField.Name(),
                    ))
                }

                foreignTable := foreignBSTag.Value()
                var ok bool
                foreignAlias, ok = self.Tables.AliasForTable(foreignTable)
                if !ok {
                    return nil, errors.Errorf(
                        "Table '%s' declared in foreign relationship is not " +
                        "included in this Query",
                        foreignTable,
                    )
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

    return columns, nil
}

func (self *Query) solveJoin() []*base.Relationship {
    results := make([]*base.Relationship, 0)
    matches := make(map[string][]base.Base)
    for _, toObject := range self.joinedObjects {
        toTableName, _ := base.BaseTable(toObject)
        if _, ok := matches[toTableName]; !ok {
            matches[toTableName] = make([]base.Base, 0)
        } else {
            continue
        }
    ToLoop:
        for _, fromObject := range self.joinedObjects {
            fromTableName, _ := base.BaseTable(fromObject)
            if _, ok := matches[fromTableName]; !ok {
                matches[fromTableName] = make([]base.Base, 0)
            }

            if objectMatches, ok := matches[fromTableName]; ok {
                for _, match := range objectMatches {
                    if match == toObject {
                        goto ToLoop
                    }
                }
            }
            if objectMatches, ok := matches[toTableName]; ok {
                for _, match := range objectMatches {
                    if match == fromObject {
                        goto ToLoop
                    }
                }
            }
            if fromObject == toObject {
                continue
            }

            joinRel, err := findRelationshipBetweenObjects(
                fromObject, toObject,
            )
            if err == nil {
                results = append(results, joinRel)
                fromTable, toTable := joinRel.Tables()
                matches[fromTable] = append(matches[fromTable], toObject)
                matches[toTable] = append(matches[toTable], fromObject)
                break
            }
        }
    }

    return results
}

// TODO: Audit performance. Consider short-circut conditions.
func findRelationshipBetweenObjects(object1, object2 base.Base) (
    *base.Relationship,
    error,
) {
    isRelationshipable := false
    matches := make(map[base.Base][]base.Relationship)

    obj1Table, _ := base.BaseTable(object1)
    obj2Table, _ := base.BaseTable(object2)

    if relationshipable, ok := object1.(base.Relationshipable); ok {
        isRelationshipable = true

        for _, relationship := range relationshipable.Relationships() {
            // TODO: Consider using reflected name to check for names as well.
            _, targetTable := relationship.Tables()
            if targetTable == obj2Table {
                if matchRels, ok := matches[object1]; ok {
                    matches[object1] = append(matchRels, relationship)
                } else {
                    matchRels := []base.Relationship{relationship}
                    matches[object1] = matchRels
                }
            }
        }
    }
    if relationshipable, ok := object2.(base.Relationshipable); ok {
        for _, relationship := range relationshipable.Relationships() {
            // TODO: Consider using reflected name to check for names as well.
            _, targetTable := relationship.Tables()
            if targetTable == obj1Table {
                if matchRels, ok := matches[object2]; ok {
                    matches[object2] = append(matchRels, relationship)
                } else {
                    matchRels := []base.Relationship{relationship}
                    matches[object2] = matchRels
                }
            }
        }
    }

    if joinRels, ok := matches[object1]; ok {
        joinRel := joinRels[0]
        return &joinRel, nil
    } else if joinRels, ok := matches[object2]; ok {
        joinRel := joinRels[0]

        return joinRel.Invert(), nil
    }

    if !isRelationshipable {
        return nil, errors.New("None of the objects have relationships")
    }

    return nil, errors.New(
        "No compatibile relationships for these two objects",
    )
}

// NewQuery returns a new Query object attached to a specified pool.
func NewQuery(pool *pool.ConnectionPool) *Query {
    aliasedTables, _ := qtypes.NewAliasedTables()
    return &Query{
        Pool: pool,
        Tables: aliasedTables,
        WhereClauses: make([]qtypes.Queryable, 0),
        SelectExpressions: make([]qtypes.SelectExpression, 0),
        typeBSFieldMap: make(map[reflect.Type]*refl.GroupedFieldsWithBS),
        typeFieldNameBSFieldMap: make(
            map[reflect.Type]*refl.GroupedFieldsWithBS,
        ),
        joinedObjects: make([]base.Base, 0),

        cached: cachedQuery{
            Select: queryValuePair{},
            From: queryValuePair{},
            Where: queryValuePair{},
            Options: queryValuePair{},
        },

        Errors: make([]error, 0),
    }
}
