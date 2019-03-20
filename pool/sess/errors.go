package sess

import (
    "errors"
)

var BaseNoIdentifierError = errors.New(
    "The base provided has no understandable identifier.",
)
var BaseIdentifierUnsetError = errors.New(
    "The base provided has an identifier but it hasn't been set.",
)
