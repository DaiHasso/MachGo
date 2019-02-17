package sess

import (
	"reflect"

	"github.com/pkg/errors"

	"github.com/daihasso/machgo/base"
	"github.com/daihasso/machgo/refl"
)

type objectIdentifier struct {
	exists,
	isSet bool
	value interface{}
}

func isEmptyValue(in interface{}) bool {
	if allowZeroValues {
		return false
	}
	result := refl.IsZeroValue(in)
	return result
}

func identifierFromBase(object base.Base) objectIdentifier {
	// TODO: Support composites.
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

	fieldGroupings := refl.GetGroupedFieldsWithBS(
		object,
		refl.GroupFieldsByTagValue("db"),
	)
	tagValueBSFields := fieldGroupings[0]

	if bsField, ok := (*tagValueBSFields)["id"]; ok {
		identifier.exists = true

		fieldName := bsField.Name()
		objVal := reflect.ValueOf(object)
		field := objVal.Elem().FieldByName(fieldName)

		if field.IsValid() {
			deepVal := field
			for deepVal.Kind() == reflect.Ptr {
				deepVal = reflect.Indirect(deepVal)
			}
			if !deepVal.IsValid() || deepVal.Interface() == nil {
				return identifier
			}

			idIn := field.Interface()
			if isEmptyValue(deepVal.Interface()) {
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

func initIdentifier(object base.Base) (*objectIdentifier, error) {
	identifier := identifierFromBase(object)
	if !identifier.exists {
		return nil, errors.New(
			"Object provided to SaveObject doesn't have an identifier.",
		)
	}

	err := checkIdentifierSet(object, &identifier)
	if err != nil {
		return nil, err
	}

	return &identifier, nil
}
