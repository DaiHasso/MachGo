package sess

import (
    "database/sql"
    "fmt"
    "math/rand"
    "sort"

    "github.com/daihasso/machgo/query/qtypes"
    "github.com/daihasso/machgo/base"
    "github.com/daihasso/machgo/refl"
    "github.com/daihasso/machgo/types"
)

func objectIdColumn(object base.Base) string {
    idColumn := "id"
    if idColumner, ok := object.(base.IDColumner); ok {
        idColumn = idColumner.IDColumn()
    }
    return idColumn
}

type sortedNamedValueIterator func(string, *sql.NamedArg)

func processSortedNamedValues(
    object base.Base, iterators ...sortedNamedValueIterator,
) {
    // TODO: Do we want to also convert non-tagged data to columns? Maybe
    //       optionally?
    var tagValues []string

    // NOTE: Is this really the best way to do this? Memory will be 2x;
    //       might not be good.
    tagValueArg := make(map[string]*sql.NamedArg)

    refl.GetFieldsByTagWithTagValues(
        object,
        "db",
        func(name string, tagValueInterface refl.TagValueInterface) {
            in := tagValueInterface.Interface
            if _, ok := in.(types.Nullable); !ok && tagValueInterface.IsNil() {
                return
            }

            tagValue := tagValueInterface.TagValue
            randomNumber := rand.Int() // #nosec: G404
            variableName := fmt.Sprintf("%s_%d", tagValue, randomNumber)
            tagValues = append(tagValues, tagValue)
            namedValue := sql.Named(variableName, in)
            tagValueArg[tagValue] = &namedValue
        },
    )

    sort.Strings(tagValues)

    for _, tagValue := range tagValues {
        namedArg := tagValueArg[tagValue]
        for _, iterator := range iterators {
            iterator(tagValue, namedArg)
        }
    }
}

func updateWhere(
    object base.Base, identifiers []base.BaseIdentifier,
) qtypes.Queryable {
    // TODO: Support composites.
    var where []qtypes.Queryable
    for _, identifier := range identifiers {
        q := qtypes.NewDefaultCondition(
            qtypes.ColumnQueryable{
                ColumnName: identifier.Column,
            },
            qtypes.InterfaceToQueryable(identifier.Value),
            qtypes.EqualCombiner,
        )

        where = append(where, q)
    }

    return qtypes.NewMultiAndCondition(where...)
}

func separateArgs(args []ObjectsOrOptions) ([]base.Base, []actionOption) {
    var (
        objects []base.Base
        options []actionOption
    )
    for _, arg := range args {
        opts, objs := arg()
        options = append(options, opts...)
        objects = append(objects, objs...)
    }

    return objects, options
}


func separateAndApply(args []ObjectsOrOptions) ([]base.Base, *actionOptions) {
    objects, options := separateArgs(args)

    optionSet := new(actionOptions)

    for _, option := range options {
        option(optionSet)
    }

    return objects, optionSet
}
