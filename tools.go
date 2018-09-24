package MachGo

import (
	"unicode"
)

// SnakeToUpperCamel converts a string in snake_case to a string in
// UpperCamelCase.
func SnakeToUpperCamel(raw string) string {
	fieldString := ""
	capitalizeNext := true
	for i := 0; i < len(raw); i++ {
		curChar := raw[i]
		if curChar == '_' {
			capitalizeNext = true
			continue
		} else if capitalizeNext {
			fieldString += string(unicode.ToUpper(rune(curChar)))
			capitalizeNext = false
		} else {
			fieldString += string(curChar)
		}
	}
	return fieldString
}
