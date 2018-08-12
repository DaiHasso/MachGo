package MachGo

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// ID is a type of attribute for an object that implements some conversion
// methods.
// Required conversions:
//   * From/To database
//   * To string
//   * From/To JSON
type ID interface {
	sql.Scanner
	driver.Valuer
	fmt.Stringer
	json.Marshaler
	json.Unmarshaler
}
