package dot

import (
	"github.com/daihasso/machgo/dsl"
)

func Query() *dsl.QuerySequence {
	return dsl.NewQuerySequence()
}
