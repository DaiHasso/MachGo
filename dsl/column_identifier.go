package dsl

import (
)

type ColumnIdentifier struct {
	columnName,
	scope string
	hasTable bool
}

func (self ColumnIdentifier) Scope() (string, bool) {
	return self.scope, self.hasTable
}

func (self ColumnIdentifier) Column() string {
	return self.columnName
}

func ColumnIdentifierFromResult(raw string) ColumnIdentifier {
	if columnAliasNamespaceRegex.MatchString(raw) {
		results := columnAliasNamespaceRegex.FindStringSubmatch(raw)

		return ColumnIdentifier{
			columnName: results[2],
			scope: results[1],
			hasTable: true,
		}
	}

	return ColumnIdentifier{
		columnName: raw,
		hasTable: false,
	}
}
