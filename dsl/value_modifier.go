package dsl

type ValueModifier func(string, string) (string, string)

func rightParenValueModifier(lh, rh string) (string, string) {
    return lh, maybeParen(rh)
}

/*
func parenValueModifier(lh, rh string) (string, string) {
    return maybeParen(lh), maybeParen(rh)
}
*/
