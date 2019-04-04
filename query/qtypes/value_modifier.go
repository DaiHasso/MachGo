package qtypes

import (
    "fmt"
)

// ValueModifier takes two statements represented by strings and does some
// mutations on them returning the mutated results.
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

