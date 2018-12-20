package refl

import (
	"unicode"
	"reflect"
)

func GuessTableName(in interface{}) string {
	t := reflect.TypeOf(in)
	t = Deref(t)
	interfaceName := t.Name()
	return UpperCamelToSnake(interfaceName)
}

func UpperCamelToSnake(raw string) string {
	fieldString := ""
	for i := 0; i < len(raw); i++ {
		curChar := raw[i]
		if unicode.IsUpper(rune(curChar)) {
			fieldString += "_"
		}
		fieldString += string(unicode.ToLower(rune(curChar)))
	}
	return fieldString
}

func LowerSnakeToUpperCamel(lowerSnake string) string {
	upperCamel := ""
	capitalizeNext := true
	for i := 0; i < len(lowerSnake); i++ {
		curChar := lowerSnake[i]
		if curChar == '_' {
			capitalizeNext = true
			continue
		} else if capitalizeNext {
			upperCamel += string(unicode.ToUpper(rune(curChar)))
			capitalizeNext = false
		} else {
			upperCamel += string(curChar)
		}
	}
	return upperCamel
}
