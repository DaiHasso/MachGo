package machgo

import (
    "database/sql/driver"
    "encoding/json"
    "fmt"
)

// IntID is an object with an integer ID.
type IntID struct {
    ID int64
}

// Scan reads data from the database into an IntID.
func (ii *IntID) Scan(src interface{}) error {
    intID, ok := src.(int64)
    if !ok {
        return fmt.Errorf(
            "Couldn't convert data from database to IntID, received "+
                "unrecognized type: %s (type: %T)",
            src,
            src,
        )
    }
    ii.ID = intID

    return nil
}

// Value returns the database value for the IntID.
func (ii IntID) Value() (driver.Value, error) {
    return ii.ID, nil
}

// String returns the string representation.
func (ii IntID) String() string {
    return fmt.Sprint(ii.ID)
}

// MarshalJSON returns the JSON representation.
func (ii IntID) MarshalJSON() ([]byte, error) {
    return json.Marshal(ii.ID)
}

// UnmarshalJSON returns the JSON representation.
func (ii IntID) UnmarshalJSON(source []byte) error {
    return json.Unmarshal(source, &(ii.ID))
}
