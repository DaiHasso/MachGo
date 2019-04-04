package dot

import (
    "github.com/daihasso/machgo/query/qtypes"
)

func Desc(q qtypes.Queryable) qtypes.Queryable {
    return qtypes.DescendingQueryable{
        Statement: q,
    }
}

func Asc(q qtypes.Queryable) qtypes.Queryable {
    return qtypes.AscendingQueryable{
        Statement: q,
    }
}
