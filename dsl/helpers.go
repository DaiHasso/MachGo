package dsl

import (
    "fmt"
)

func maybeParen(s string) string {
    if s[0] == '(' {
        return s
    }

    return fmt.Sprintf("(%s)", s)
}
