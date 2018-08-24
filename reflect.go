package MachGo

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

// GetObjectSlice For some arbitrary object get an interface that
// represents a slice of its type.
func GetObjectSlice(obj Object) interface{} {
	ptr := reflect.New(reflect.SliceOf(reflect.TypeOf(obj)))
	iface := ptr.Interface()

	return iface
}
