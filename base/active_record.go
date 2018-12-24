package base

import (
	"reflect"

	"github.com/pkg/errors"

	"MachGo"
	"MachGo/dsl"
)

type ActiveRecorder interface {
	Save() error
	Update() error
	Delete() error
	Get(interface{}) error
	Find(...dsl.Queryable) (MachGo.QueryResults, error)
	FindAll() (MachGo.QueryResults, error)
	LinkInstance(Base)
}

type ActiveRecord struct {
	instanceRef Base
}

// TODO: Delete this when all the methods are implemented.
var unimplementedError = "This method hasn't been implemented"

func (self ActiveRecord) Save() error {
	// TODO: Implement.
	return errors.New(unimplementedError)
}
func (self ActiveRecord) Update() error {
	// TODO: Implement.
	return errors.New(unimplementedError)
}
func (self ActiveRecord) Delete() error {
	// TODO: Implement.
	return errors.New(unimplementedError)
}
func (self ActiveRecord) Get(interface{}) error {
	// TODO: Implement.
	return errors.New(unimplementedError)
}
func (self ActiveRecord) Find(...dsl.Queryable) (MachGo.QueryResults, error) {
	// TODO: Implement.
	return MachGo.QueryResults{}, errors.New(unimplementedError)
}
func (self ActiveRecord) FindAll() (MachGo.QueryResults, error) {
	// TODO: Implement.
	return MachGo.QueryResults{}, errors.New(unimplementedError)
}
func (self ActiveRecord) LinkInstance(ref Base) {
	self.instanceRef = ref
}

// FIXME: Better name pl0x.
func typeBase(arType reflect.Type) ObjectActiveRecordLinker {
	return func(object Base) error {
		objType := reflect.TypeOf(object)
		if objType.Kind() != reflect.Ptr ||
			objType.Elem().Kind() == reflect.Ptr {
			return errors.Errorf(
				"Object provided must be a pointer to an object not an " +
					"object by value, pointer to a pointer or so on. " +
					"Provided type: %T",
				object,
			)
		}
		arInType := reflect.TypeOf((*ActiveRecorder)(nil)).Elem()
		objVal := reflect.ValueOf(object)
		arField := objVal.Elem().FieldByName("ActiveRecorder")

		if !arField.IsValid() {
			return errors.New(
				"The 'ActiveRecorder' field does not exist on the object " +
					"provided, did you forget to embedd the ActiveRecord " +
					"interface?",
			)
		}

		if arField.Type() != arInType {
			return errors.Errorf(
				"The 'ActiveRecorder' field on the provided object exists " +
					"but is not of the ActiveRecorder type provided to the " +
					"ActiveRecordLinker: %s",
				arType.String(),
			)
		}

		newAr := reflect.New(arType)
		newArIn := newAr.Elem().Interface()
		newArInCoerced := newArIn.(ActiveRecorder)
		newArInCoerced.LinkInstance(object)

		arField.Set(reflect.ValueOf(newArInCoerced))

		return nil
	}
}

type ObjectActiveRecordLinker func(object Base) error

var LinkActiveRecord = typeBase(reflect.TypeOf((*ActiveRecord)(nil)).Elem())

func NewActiveRecordLinker(ar ActiveRecorder) ObjectActiveRecordLinker {
	arType := reflect.TypeOf(ar)

	return typeBase(arType)
}
