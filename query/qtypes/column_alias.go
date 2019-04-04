package qtypes

import (
    "fmt"
    "regexp"
)

var columnAliasStringRegex = regexp.MustCompile(`^([^_]+)_(.*)$`)

// ColumnAlias is a table's alias paired with a column in a standardized
// format.
type ColumnAlias struct {
    TableAlias,
    ColumnName string
}

type ColumnAliasField struct {
    ColumnAlias
    FieldName string
}

func (self ColumnAlias) String() string {
    return fmt.Sprintf("%s_%s", self.TableAlias, self.ColumnName)
}

// ColumnAliasForString splits a string in the format of alias_column into a
// ColumnAlias type.
func ColumnAliasFromString(rawColumn string) (*ColumnAlias, bool) {
    if !columnAliasStringRegex.MatchString(rawColumn) {
        return nil, false
    }

    results := columnAliasStringRegex.FindStringSubmatch(rawColumn)

    columnAlias := &ColumnAlias{
        TableAlias: results[1],
        ColumnName: results[2],
    }

    return columnAlias, true
}
