package refl

import (
	"strings"
)

type BSTag struct {
	name,
	value string
	properties []string
}

func (self BSTag) Value() string {
	return self.value
}

func (self BSTag) Name() string {
	return self.name
}

func (self BSTag) AllProperties() []string {
	return self.properties
}

func (self BSTag) HasProperty(key string) bool {
	for _, propertyString := range self.properties {
		if propertyString == key {
			return true
		}
	}
	return false
}

func (self BSTag) Property(i int) string {
	return self.properties[i]
}

func propertiesFromValue(rawValue string) (string, []string) {
	var value string
	splitStrings := strings.Split(rawValue, ",")
	value = splitStrings[0]
	var properties []string
	if len(splitStrings) > 1 {
		rawProps := splitStrings[1:]
		properties = make([]string, len(rawProps))
		for i, s := range rawProps {
			cleanedProp := strings.Trim(s, " ")
			properties[i] = cleanedProp
		}
	}

	return value, properties
}

func newBSTag(name, rawValue string) *BSTag {
	value, properties := propertiesFromValue(rawValue)
	return &BSTag{
		name: name,
		value: value,
		properties: properties,
	}
}
