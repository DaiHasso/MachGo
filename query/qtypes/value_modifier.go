package qtypes

import (
    "fmt"
)

type ValueModifier func(string, string) (string, string)

func maybeParen(s string) string {
    if s[0] == '(' {
        return s
    }

    return fmt.Sprintf("(%s)", s)
}

func rightParenValueModifier(lh, rh string) (string, string) {
    return lh, maybeParen(rh)
}

