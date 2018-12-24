package base

import (
)

// LiteralStatement is a string that will be directly inserted into a
// statement. Use great caution when using this as it will be vunerable to
// injection if not used properly.
type LiteralStatement string
