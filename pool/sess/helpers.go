package sess

import (
	"database/sql"
	"fmt"
	"sort"

	"MachGo/base"
	"MachGo/refl"
)

func objectIdColumn(object base.Base) string {
	idColumn := "id"
	if idColumner, ok := object.(base.IDColumner); ok {
		idColumn = idColumner.IDColumn()
	}
	return idColumn
}

type sortedNamedValueIterator func(string, *sql.NamedArg)

func processSortedNamedValues(
	object base.Base, iterators ...sortedNamedValueIterator,
) {
	// TODO: Do we want to also convert non-tagged data to columns? Maybe
	//       optionally?
	var variableNames []string

	// NOTE: Is this really the best way to do this? Memory will be 2x;
	//       might not be good.
	tagValueArg := make(map[string]*sql.NamedArg)

	refl.GetFieldsByTagWithTagValues(
		object,
		"db",
		func(name string, tagValueInterface refl.TagValueInterface) {
			tagValue := tagValueInterface.TagValue
			variableNames = append(variableNames, tagValue)
			namedValue := sql.Named(tagValue, tagValueInterface.Interface)
			tagValueArg[tagValue] = &namedValue
		},
	)

	sort.Strings(variableNames)

	for _, variableName := range variableNames {
		namedArg := tagValueArg[variableName]
		for _, iterator := range iterators {
			iterator(variableName, namedArg)
		}
	}
}

func updateWhere(
	object base.Base, identifier objectIdentifier,
) (string, []interface{}) {
	// TODO: Support composites.
	whereString := ""
	idColumn := objectIdColumn(object)

	bindvar := "identifier"
	namedIdentifier := sql.Named(bindvar, identifier.value)

	whereString = fmt.Sprintf("%s = @%s", idColumn, bindvar)

	return whereString, []interface{}{namedIdentifier}
}

func separateArgs(args []ObjectOrOption) ([]base.Base, []actionOption) {
	var (
		objects []base.Base
		options []actionOption
	)
	for _, arg := range args {
		if opt, ok := arg.(actionOption); ok {
			options = append(options, opt)
		} else {
			objects = append(objects, arg)
		}
	}

	return objects, options
}


func separateAndApply(args []ObjectOrOption) ([]base.Base, *actionOptions) {
	objects, options := separateArgs(args)

	optionSet := new(actionOptions)

	for _, option := range options {
		option(optionSet)
	}

	return objects, optionSet
}
