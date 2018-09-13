package refl

import (
	"reflect"
)

type FieldWithBS struct {
	name string
	fieldValue *reflect.Value
	nameTagMap map[string]*BSTag
	tags []*BSTag
}

func (self FieldWithBS) Tags() []*BSTag {
	return self.tags
}

func (self FieldWithBS) Tag(key string) *BSTag {
	if bsTag, ok := self.nameTagMap[key]; ok {
		return bsTag
	}
	return nil
}

func (self FieldWithBS) Interface() interface{} {
	return self.fieldValue.Interface()
}

func (self FieldWithBS) Name() string {
	return self.name
}

func newFieldWithBS(
	name string,
	fieldValue *reflect.Value,
	nameTagMap map[string]*BSTag,
) *FieldWithBS {
	tags := make([]*BSTag, len(nameTagMap))

	i := 0
	for _, tag := range nameTagMap {
		tags[i] = tag
		i++
	}

	return &FieldWithBS{
		name: name,
		fieldValue: fieldValue,
		nameTagMap: nameTagMap,
		tags: tags,
	}
}
