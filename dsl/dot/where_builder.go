package dot

import (
	"MachGo/dsl"
)

// Where returns a new WhereBuilder.
func Where(conditions ...dsl.WhereConditioner) *dsl.WhereBuilder {
	whereBuilder := dsl.NewWhere()
	for _, condition := range conditions {
		condition(&whereBuilder.Conditions)
	}
	return whereBuilder
}
