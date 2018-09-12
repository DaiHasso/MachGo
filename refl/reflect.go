package refl

import (
	"fmt"
	"reflect"
	"unicode"
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

// GetFieldsByTagWithTagValues will get all fields in an object for a
// given tag and also the values associated with the specified tag.
func GetFieldsByTagWithTagValues(
	in interface{},
	tag string,
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
			value := v.Field(i).Interface()
			tagValueInterface := TagValueInterface{tagValue, value}
			nameTagValueInterfaces[fieldName] = tagValueInterface
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
func GetTagValues(in interface{}, tag string, formatters... (func(string) string)) []string {
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
