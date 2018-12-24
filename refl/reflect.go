package refl

import (
	"fmt"
	"reflect"
	"strconv"
	"unicode"

	"github.com/pkg/errors"

	logging "github.com/daihasso/slogging"
)

// TagValueInterface is a type which holds both a TagValue and an
// interface.
type TagValueInterface struct {
	TagValue  string
	Interface interface{}
}

// IsUnset checks if the value of the TagValueInterface is unset.
func (tvi *TagValueInterface) IsUnset() bool {
	zeroType := reflect.Zero(reflect.TypeOf(tvi.Interface)).Interface()
	return reflect.DeepEqual(tvi.Interface, zeroType)
}

type fieldTagIterator func(string, TagValueInterface)

// GetFieldsByTagWithTagValues will get all fields in an object for a
// given tag and also the values associated with the specified tag.
// TODO: This should probably return an iterator or something so that the
//       resulting code doesn't repeat the iteration.
func GetFieldsByTagWithTagValues(
	in interface{},
	tag string,
	iterators ...fieldTagIterator,
) map[string]TagValueInterface {
	v := reflect.ValueOf(in)
	if v.Kind() == reflect.Ptr {
		in = reflect.Indirect(v).Interface()
		v = reflect.ValueOf(in)
	}

	t := reflect.TypeOf(in)

	numFields := v.NumField()
	nameTagValueInterfaces := make(
		map[string]TagValueInterface,
		numFields,
	)

	for i := 0; i < numFields; i++ {
		field := t.Field(i)
		fieldName := field.Name

		if !unicode.IsUpper(rune(fieldName[0])) {
			continue
		}

		if tagValue, ok := field.Tag.Lookup(tag); ok {
			bsTagValue := newBSTag(tag, tagValue)
			if bsTagValue.HasProperty("foreign") {
				// TODO: Handle this more elegantly.
				continue
			}
			value := v.Field(i).Interface()
			tagValueInterface := TagValueInterface{bsTagValue.Value(), value}
			nameTagValueInterfaces[fieldName] = tagValueInterface

			for _, iterator := range iterators {
				iterator(fieldName, tagValueInterface)
			}
		}
	}

	return nameTagValueInterfaces
}

// GetFieldsByTag will get all fields in an object for a given tag.
func GetFieldsByTag(in interface{}, tag string) map[string]interface{} {
	t := reflect.TypeOf(in)
	v := reflect.ValueOf(in)

	numFields := v.NumField()
	nameValues := make(map[string]interface{}, numFields)

	for i := 0; i < numFields; i++ {
		field := t.Field(i)

		if _, ok := field.Tag.Lookup(tag); ok {
			value := v.Field(i).Interface()
			nameValues[field.Name] = value
		}
	}

	return nameValues
}

// GetTagValues will get all the fields tag values as a slice.
func GetTagValues(
	in interface{}, tag string, formatters... (func(string) string),
) []string {
	t := reflect.TypeOf(in)
	v := reflect.ValueOf(in)
	for t.Kind() == reflect.Ptr {
		t = Deref(t)
		v = v.Elem()
	}

	numFields := v.NumField()
	tagValues := make([]string, 0)

	for i := 0; i < numFields; i++ {
		field := t.Field(i)

		if tagVal, ok := field.Tag.Lookup(tag); ok {
			if len(formatters) != 0 {
				for _, formatter := range formatters {
					tagVal = formatter(tagVal)
				}
			}
			tagValues = append(tagValues, tagVal)
		}
	}

	return tagValues
}

// InterfaceToSQLValue will take an interface, and try to cleverly convert it
// to the best SQL value representation.
func InterfaceToSQLValue(
	in interface{},
) string {

	switch reflect.TypeOf(in).Kind() {
	case
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64:
		return fmt.Sprintf("%v", in)
	case reflect.String:
		return fmt.Sprintf("'%v'", in)
	}
	return ""
}

// GetInterfaceName will get the resolved name of an interface.
func GetInterfaceName(in interface{}) string {
	resolvedName := ""
	t := reflect.TypeOf(in)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
		resolvedName += "*"
	}
	return resolvedName + t.Name()
}

// GetInterfaceSlice for some arbitrary interface get an interface that
// represents a slice of its type.
func GetInterfaceSlice(in interface{}) interface{} {
	slice := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(in)), 0, 0)
	slicePtr := reflect.New(slice.Type())
	slicePtr.Elem().Set(slice)

	return slicePtr.Interface()
}

// Deref is Indirect for reflect.Types
// Pulled from sqlx/reflectx: https://tinyurl.com/y9akfp5n
func Deref(t reflect.Type) reflect.Type {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// DerefDeep derefs all the way down.
func DerefDeep(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// GroupedFieldsWithBS is a map of some grouping string to a FieldWithBS.
type GroupedFieldsWithBS map[string]*FieldWithBS

// GroupFieldWithBSBy takes in a new FieldWithBS and the existing
// GroupedFieldsWithBS and inserts the FieldWithBs appropriately according to
// its grouping function.
type GroupFieldWithBSBy func(*GroupedFieldsWithBS, *FieldWithBS)

// GetGroupedFieldsWithBS takes an interface and returns a series of maps that
// group FieldWithBS by the provided grouping functions in provided order.
func GetGroupedFieldsWithBS(
	in interface{},
	byFilters ...GroupFieldWithBSBy,
) []*GroupedFieldsWithBS {
	v := reflect.ValueOf(in)
	for v.Kind() == reflect.Ptr {
		in = reflect.Indirect(v).Interface()
		v = reflect.ValueOf(in)
	}

	t := reflect.TypeOf(in)

	numFields := v.NumField()
	allGroupings := make([]*GroupedFieldsWithBS, len(byFilters))
	for i := range allGroupings {
		fieldsWithBS := make(GroupedFieldsWithBS, numFields)
		allGroupings[i] = &fieldsWithBS
	}

	for i := 0; i < numFields; i++ {
		field := t.Field(i)
		fieldName := field.Name

		if !unicode.IsUpper(rune(fieldName[0])) {
			continue
		}

		fieldValue := v.Field(i)
		tagValues, err := getAllTags(field.Tag)
		if err != nil {
			logging.Debug("Corrupt tag encountered.").Send()
		}
		fieldWithBS := newFieldWithBS(fieldName, &fieldValue, tagValues)
		for i, byFilter := range byFilters {
			byFilter(allGroupings[i], fieldWithBS)
		}
	}

	return allGroupings
}

// GroupFieldsByTagValue is a grouping function that will group all the input
// FieldWithBS by its tag value.
func GroupFieldsByTagValue(
	tagNames ...string,
) GroupFieldWithBSBy {
	fn := func(
		fieldsWithBS *GroupedFieldsWithBS,
		fieldWithBS *FieldWithBS,
	) {
		for _, tagName := range tagNames {
			bsTag := fieldWithBS.Tag(tagName)
			if bsTag == nil {
				return
			}
			tagValue := bsTag.Value()
			(*fieldsWithBS)[tagValue] = fieldWithBS
		}
	}
	return fn
}

// GroupFieldsByFieldName is a grouping function that will group all the input
// FieldWithBS by its field name.
func GroupFieldsByFieldName() GroupFieldWithBSBy {
	fn := func(fieldsWithBS *GroupedFieldsWithBS, fieldWithBS *FieldWithBS) {
		fieldName := fieldWithBS.Name()
		(*fieldsWithBS)[fieldName] = fieldWithBS
	}
	return fn
}

// GroupFieldsByTagName is a grouping function that will group all the input
// FieldWithBS by its tag name.
func GroupFieldsByTagName() GroupFieldWithBSBy {
	fn := func(fieldsWithBS *GroupedFieldsWithBS, fieldWithBS *FieldWithBS) {
		for _, bsTag := range fieldWithBS.Tags() {
			(*fieldsWithBS)[bsTag.Name()] = fieldWithBS
		}
	}
	return fn
}

func getAllTags(
	tag reflect.StructTag,
) (map[string]*BSTag, error) {
	tagBSTags := make(map[string]*BSTag)
	// Heavily influenced by https://tinyurl.com/ycoqx5hn

	curTagPart := tag
	for curTagPart != "" {
		// Skip leading space.
		i := 0
		for i < len(curTagPart) && curTagPart[i] == ' ' {
			i++
		}

		curTagPart = curTagPart[i:]
		if curTagPart == "" {
			break
		}


		// Scan to colon. A space, a quote or a control character is a syntax
		// error. Strictly speaking, control chars include the range
		// [0x7f, 0x9f], not just [0x00, 0x1f], but in practice, we ignore the
		// multi-byte control characters as it is simpler to inspect the tag's
		// bytes than the tag's runes.
		i = 0
		for i < len(curTagPart) {
			curRune := curTagPart[i]
			if curRune <= ' ' || curRune == ':' || curRune == '"' ||
				curRune == 0x7f {
				break
			}
			i++
		}

		if i == 0 || i+1 >= len(curTagPart) ||
			curTagPart[i] != ':' || curTagPart[i+1] != '"' {
			break
		}

		name := string(curTagPart[:i])

		curTagPart = curTagPart[i+1:]


		// Scan quoted string to find value.
		i = 1
		for i < len(curTagPart) && curTagPart[i] != '"' {
			if curTagPart[i] == '\\' {
				i++
			}
			i++
		}

		if i >= len(curTagPart) {
			break
		}

		qvalue := string(curTagPart[:i+1])
		curTagPart = curTagPart[i+1:]

		value, err := strconv.Unquote(qvalue)
		if err != nil {
			return tagBSTags, err
		}
		tagBSTags[name] = newBSTag(name, value)
	}

	return tagBSTags, nil
}

// ElementTypeFromSlice takes a slice or a pointer to slice and returns the
// type of its element(s).
func ElementTypeFromSlice(sliceIn interface{}) (reflect.Type, error) {
	sliceVal := reflect.ValueOf(sliceIn)

    for sliceVal.Kind() == reflect.Ptr {
		sliceVal = sliceVal.Elem()
	}

	if sliceVal.Kind() != reflect.Slice {
		return nil, errors.Errorf(
			"Argumnt to ElementTypeFromSlice must be a Slice or a pointer to "+
				"a Slice not '%T'.",
			sliceIn,
		)
	}

	sliceType := sliceVal.Type()
	elemType := sliceType.Elem()

	return elemType, nil
}

func InitSetField(fieldVal, newVal reflect.Value) error {
	derefedNewVal := newVal
	for derefedNewVal.Kind() == reflect.Ptr {
		derefedNewVal = derefedNewVal.Elem()
	}

	if fieldVal.Kind() == reflect.Ptr && fieldVal.IsNil() {
		alloc := reflect.New(Deref(fieldVal.Type()))
		fieldVal.Set(alloc)
	}

	derefedFieldVal := fieldVal
	for derefedFieldVal.Kind() == reflect.Ptr {
		derefedFieldVal = derefedFieldVal.Elem()
	}

	switch k := derefedFieldVal.Kind(); k {
		// TODO: Add more handled kinds.
		case reflect.Int64:
		intValue := derefedNewVal.Interface().(int64)
		derefedFieldVal.SetInt(intValue)
		default:
		return errors.Errorf("Unknown kind: %s", k.String())
	}

	return nil
}
