package refl

import (
	"strings"
)

type BSTag struct {
	name,
	value string
	properties map[string]bool
}

func (self BSTag) Value() string {
	return self.value
}

func (self BSTag) Name() string {
	return self.name
}

func (self BSTag) AllProperties() map[string]bool {
	return self.properties
}

func (self BSTag) Property(key string) bool {
	_, hasProperty := self.properties[key]
	return hasProperty
}

func propertiesFromValue(rawValue string) (string, map[string]bool) {
	var value string
	var properties map[string]bool
	splitStrings := strings.Split(rawValue, ",")
	value = splitStrings[0]
	if len(splitStrings) > 1 {
		rawProps := splitStrings[1:]
		properties = make(map[string]bool, len(rawProps)-1)
		for _, s := range rawProps {
			cleanedProp := strings.Trim(s, " ")
			properties[cleanedProp] = true
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
