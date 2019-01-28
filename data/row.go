package data

import (
	"errors"
)

// Row represents a materialized record.
type Row struct {
	Index  Index
	Values []Observation
}

// Rows are materialized table data, suitable for printing for example.
type Rows struct {
	Data   []Row
	Schema Schema
}

// Observation accesses a column value by name instead of column index.
func (r Row) Observation(c ColumnName) (Observation, error) {
	// TODO more efficiently access schema
	for _, o := range r.Values {
		if o.ColumnName() == c {
			return o, nil
		}
	}
	return Observation{}, errors.New("no column " + string(c))
}

// Observation holds an arbitrary, nullable column value.
type Observation struct {
	col   *ColumnName
	value interface{}
}

// ColumnName returns the column name for the value.
func (o Observation) ColumnName() ColumnName {
	return *o.col
}

// IsNull returns true if the value is null.
func (o Observation) IsNull() bool {
	return o.value == nil
}

// Interface returns the underlying representation of the value.
func (o Observation) Interface() interface{} {
	return o.value
}
