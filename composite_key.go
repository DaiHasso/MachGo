package machgo

// CompositeKey is a composite key made up of one or more columns.
type CompositeKey interface {
    GetColumnNames() []string
    SetColumnNames([]string)
}

// SimpleCompositeKey is a simple composite key implementation.
type SimpleCompositeKey struct {
    columns []string
}

// GetColumnNames retrieves the database column names for the composite key.
func (sck SimpleCompositeKey) GetColumnNames() []string {
    return sck.columns
}

// SetColumnNames retrieves the database column names for the composite key.
func (sck *SimpleCompositeKey) SetColumnNames(columns []string) {
    sck.columns = columns
}
