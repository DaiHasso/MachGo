package sess

import (
	"database/sql"
	"reflect"
	"sort"

	"github.com/pkg/errors"

	"MachGo/base"
	"MachGo/refl"
)

type objectIdentifier struct {
	exists,
	isSet bool
	value interface{}
}

func genericIDMatcher(s string) bool {
	// TODO: Are there any other cases?

	if s == "Id" || s == "ID" {
		return true
	}

	return false
}

func identifierFromBase(object base.Base) objectIdentifier {
	identifier := objectIdentifier{
		exists: false,
		isSet: false,
		value: nil,
	}

	if v, ok := object.(base.TraditionalID); ok {
		identifier.exists = true
		id, ok := v.ID()
		if !ok {
			return identifier
		}
		identifier.isSet = true
		identifier.value = id
		return identifier
	}

	// Fallback on pulling it from the struct directly.
	value := reflect.Indirect(reflect.ValueOf(object))
	matched := false
	identifierValue := value.FieldByNameFunc(func(s string) bool{
		result := genericIDMatcher(s)
		matched = matched || result
		return result
	})
	if matched {
		identifier.exists = true
		if identifierValue.IsValid() {
			idIn := identifierValue.Interface()
			deepVal := identifierValue
			for deepVal.Kind() == reflect.Ptr {
				deepVal = reflect.Indirect(deepVal)
			}
			if !deepVal.IsValid() || deepVal.Interface() == nil {
				return identifier
			}

			identifier.isSet = true
			identifier.value = idIn
		}
	}

	return identifier
}

func setIdentifierOnBase(
	object base.Base, newIdValue interface{},
) error {
	if v, ok := object.(base.TraditionalID); ok {
		return v.SetID(newIdValue)
	}
	fieldGroupings := refl.GetGroupedFieldsWithBS(
		object,
		refl.GroupFieldsByTagValue("db"),
	)
	tagValueBSFields := fieldGroupings[0]
	fieldName := (*tagValueBSFields)["id"].Name()
	objVal := reflect.ValueOf(object)
	field := objVal.Elem().FieldByName(fieldName)
	if !field.IsValid() {
		return errors.Errorf(
			"Field in returned data '%s' is not valid.",
			fieldName,
		)
	}
	
	newValue := reflect.ValueOf(newIdValue)

	return refl.InitSetField(field, newValue)
}

func checkIdentifierSet(
	object base.Base, identifier *objectIdentifier,
) error {
	idGenerator, isGenerator := object.(base.IDGenerator)
	_, isTraditional := object.(base.TraditionalID)
	idVal := reflect.ValueOf(identifier)

	if isGenerator {
		if !identifier.isSet {
			if (idVal.Kind() != reflect.Ptr && !isTraditional) {
				return errors.New(
					"Can't use an IDGenerator without id being a pointer or " +
						"implementing TraditionID.",
				)
			}

			id := idGenerator.NewID()
			err := setIdentifierOnBase(object, id)
			if err != nil {
				return err
			}

			identifier.isSet = true
		}
	}

	return nil
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
