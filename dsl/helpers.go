package dsl

import (
	"fmt"
	"unicode"
)

func maybeParen(s string) string {
	if s[0] == '(' {
		return s
	}

	return fmt.Sprintf("(%s)", s)
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
