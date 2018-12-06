package dot

import (
	"MachGo/dsl"
)

func Eq(a, b dsl.WhereValuer) dsl.WhereConditioner {
	return Equal(a, b)
}

func Equal(a, b dsl.WhereValuer) dsl.WhereConditioner {
	return dsl.WhereEqual(a, b)
}

func Gt(a, b dsl.WhereValuer) dsl.WhereConditioner {
	return Greater(a,b)
}

func Greater(a, b dsl.WhereValuer) dsl.WhereConditioner {
	return dsl.WhereGreater(a, b)
}

func Lt(a, b dsl.WhereValuer) dsl.WhereConditioner {
	return Less(a, b)
}

func Less(a, b dsl.WhereValuer) dsl.WhereConditioner {
	return dsl.WhereLess(a, b)
}

func Gte(a, b dsl.WhereValuer) dsl.WhereConditioner {
	return GreaterEqual(a, b)
}

func GreaterEqual(a, b dsl.WhereValuer) dsl.WhereConditioner {
	return dsl.WhereGreaterEqual(a, b)
}

func Lte(a, b dsl.WhereValuer) dsl.WhereConditioner {
	return LessEqual(a, b)
}

func LessEqual(a, b dsl.WhereValuer) dsl.WhereConditioner {
	return dsl.WhereLessEqual(a, b)
}

func In(a, b dsl.WhereValuer) dsl.WhereConditioner {
	return dsl.WhereIn(a, b)
}
