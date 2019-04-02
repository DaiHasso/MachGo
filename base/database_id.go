package base

import (
    "reflect"
   
    "github.com/pkg/errors"

    "github.com/daihasso/machgo/refl"
)

var allowZeroValues = false;

type DatabaseIDGenerator interface {
    DatabaseGeneratedID()
}

type DatabaseFuncIDGenerator interface {
    DatabaseIDGenerationFunc() LiteralStatement
}

type DatabaseManagedID struct {}
func (DatabaseManagedID) DatabaseGeneratedID() {}

type BaseIdentifier struct {
    Column string
    Exists,
    IsSet bool
    Value interface{}
}

func GetId(object Base) []BaseIdentifier {
    return identifierFromBase(object)
}

func SetId(object Base, newIdValue interface{}) error {
    return setIdentifierOnBase(object, newIdValue)
}

func InitializeId(object Base) ([]BaseIdentifier, error) {
    return initIdentifier(object)
}

func isEmptyValue(in interface{}) bool {
    if allowZeroValues {
        return false
    }
    result := refl.IsZeroValue(in)
    return result
}

func identifierFromColumn(
    tagValueBSFields refl.GroupedFieldsWithBS,
    object Base,
    columnName string,
) BaseIdentifier {
    identifier := BaseIdentifier{
        Column: columnName,
        Exists: false,
        IsSet: false,
        Value: nil,
    }

    if bsField, ok := tagValueBSFields[columnName]; ok {
        identifier.Exists = true

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

            identifier.IsSet = true
            identifier.Value = idIn
        }
    }

    return identifier
}

func identifierFromBase(object Base) []BaseIdentifier {
    // TODO: Support composites.
    if v, ok := object.(TraditionalID); ok {
        identifier := BaseIdentifier{
            Exists: false,
            IsSet: false,
            Value: nil,
        }

        identifier.Exists = true
        id, ok := v.ID()
        if !ok {
            return nil
        }
        identifier.IsSet = true
        identifier.Value = id
        return []BaseIdentifier{identifier}
    }

    columns := []string{"id"}
    if composite, ok := object.(CompositeKey); ok {
        columns = composite.CompositeKey()
    }

    fieldGroupings := refl.GetGroupedFieldsWithBS(
        object,
        refl.GroupFieldsByTagValue("db"),
    )
    tagValueBSFields := fieldGroupings[0]

    var identifiers []BaseIdentifier
    for _, columnName := range columns {
        identifier := identifierFromColumn(
            *tagValueBSFields, object, columnName,
        )
        identifiers = append(identifiers, identifier)
    }

    return identifiers
}

func setIdentifierOnBase(object Base, newIdValue interface{}) error {
    if v, ok := object.(TraditionalID); ok {
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

func checkIdentifierSet(object Base, identifier *BaseIdentifier) error {
    idGenerator, isGenerator := object.(IDGenerator)
    _, isTraditional := object.(TraditionalID)
    idVal := reflect.ValueOf(identifier)

    if isGenerator {
        if !identifier.IsSet {
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

            identifier.IsSet = true
        }
    }

    return nil
}

func initIdentifier(object Base) ([]BaseIdentifier, error) {
    identifiers := identifierFromBase(object)
    for i, identifier := range identifiers {
        if !identifier.Exists {
            return nil, errors.New(
                "Object provided to SaveObject doesn't have an identifier.",
            )
        }

        err := checkIdentifierSet(object, &identifier)
        if err != nil {
            return nil, err
        }
        identifiers[i] = identifier
    }

    return identifiers, nil
}
