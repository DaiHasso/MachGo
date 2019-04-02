package sess

import (
    "github.com/daihasso/machgo/base"
)

type actionOptions struct {
    stopOnFailure bool
}

type actionOption func(*actionOptions)

var StopOnFailure = func() ([]actionOption, []base.Base) {
    return []actionOption{
        func(ops *actionOptions) {
            ops.stopOnFailure = true
        },
    }, nil
}

func Objs(objs ...base.Base) ObjectsOrOptions {
    return func() ([]actionOption, []base.Base) {
        return nil, objs
    }
}

type ObjectsOrOptions func() ([]actionOption, []base.Base)
