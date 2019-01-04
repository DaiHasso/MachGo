package sess

import (
	"database/sql"
	"fmt"
	"strings"

	"MachGo/base"
)

var (
	bindvarReplacers = []string{
		"bindvars", "bindvar",
	}
	columnNameReplacers = []string{
		"columnNames", "columnnames", "columnName", "columnname", "columns",
	}
)

type QueryParts struct {
	Bindvars,
	ColumnNames string
	VariableValues []interface{}
}

func (QueryParts) appendParts(existing *string, parts []string) {
	if len(*existing) > 0 {
		*existing += ", "
	}

	*existing += strings.Join(parts, ", ")
}

func (self *QueryParts) AddBindvar(bindvars ...string) {
	self.appendParts(&self.Bindvars, bindvars)
}

func (self *QueryParts) AddColumnName(columnNames ...string) {
	self.appendParts(&self.ColumnNames, columnNames)
}

func (self *QueryParts) AddValue(values ...interface{}) {
	self.VariableValues = append(self.VariableValues, values...)
}

func (self QueryParts) Format(template string, extras ...interface{}) string {
	// TODO: Using replacer here might not be the most efficient method.
	var pairs []string
	for _, replacementString := range bindvarReplacers {
		pairs = append(
			pairs,
			[]string{"{" + replacementString + "}", self.Bindvars}...,
		)
	}
	for _, replacementString := range columnNameReplacers {
		pairs = append(
			pairs,
			[]string{"{" + replacementString + "}", self.ColumnNames}...,
		)
	}
	replacer := strings.NewReplacer(pairs...)
	return fmt.Sprintf(replacer.Replace(template), extras...)
}

type ColumnFilter func(string, *sql.NamedArg) bool

func QueryPartsFromObject(
	object base.Base, filters ...ColumnFilter,
) QueryParts {
	queryParts := new(QueryParts)
	processSortedNamedValues(
		object, func(columnName string, namedArg *sql.NamedArg) {
			for _, filter := range filters {
				if filter(columnName, namedArg) {
					return
				}
			}

			queryParts.AddColumnName(columnName)
			queryParts.AddBindvar("@" + columnName)
			queryParts.AddValue(*namedArg)
		},
	)

	return *queryParts
}
