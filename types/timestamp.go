package types

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// Timestamp is a time.Time that marshals to unix timestamp.
type Timestamp struct {
	time.Time
}

// MarshalJSON will marshal the timestamp into a unix timestamp.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	unixTimestamp := fmt.Sprintf("%v", t.Unix())
	return []byte(unixTimestamp), nil
}

// UnmarshalJSON will read a timestamp from json.
func (t *Timestamp) UnmarshalJSON(source []byte) error {
	i, err := strconv.ParseInt(string(source), 10, 64)
	if err != nil {
		return err
	}
	parsedTime := time.Unix(i, 0)

	t.Time = parsedTime

	return nil
}

// Value gets the value of a Timestamp for writing to the DB.
func (t *Timestamp) Value() (driver.Value, error) {
	return t.Time, nil
}

// Scan will read a value into a new Timestamp.
func (t *Timestamp) Scan(src interface{}) error {
	var source time.Time

	switch src.(type) {
	case time.Time:
		source = src.(time.Time)
	default:
		return errors.New("Incompatible type for Timestamp")
	}

	*t = Timestamp{source}

	return nil
}

func (t Timestamp) String() string {
	return t.Time.String()
}
