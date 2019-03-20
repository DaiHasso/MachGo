package refl

import (
    "unicode"
    "reflect"

    "github.com/pkg/errors"
)

func GuessTableName(in interface{}) (string, error) {
    t := reflect.TypeOf(in)
    t = Deref(t)
    interfaceName := t.Name()
    if len(interfaceName) == 0 {
        return "", errors.New(
            "Couldn't determine table name because object provided has no " +
                "struct name.",
        )
    }
    tableName := UpperCamelToSnake(interfaceName)
    if tableName[len(tableName)-1] != 's' {
        tableName += "s"
    }
    return tableName, nil
}

func UpperCamelToSnake(raw string) string {
    fieldString := ""
    for i := 0; i < len(raw); i++ {
        curChar := raw[i]
        if unicode.IsUpper(rune(curChar)) && i != 0{
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
