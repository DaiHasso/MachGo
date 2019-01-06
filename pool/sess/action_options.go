package sess

import (
)

type actionOptions struct {
	stopOnFailure bool
}

type actionOption func(*actionOptions)

var StopOnFailure = func(ops *actionOptions) {
	ops.stopOnFailure = true
}

type ObjectOrOption interface{}
