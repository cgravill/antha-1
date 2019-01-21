package data

import (
	"errors"
)

// Row represents a materialized record. Every time you make a row you are
// copying data so this should be minimized
type Row struct {
	Index  Index
	Values []Observation
}

// Rows are materialized table data
type Rows struct {
	Data   []Row
	Schema Schema
}

func (r Row) Observation(c ColumnName) (Observation, error) {
	// TODO more efficiently access schema
	// TODO access value by index
	for _, o := range r.Values {
		if o.ColumnName() == c {
			return o, nil
		}
	}
	return Observation{}, errors.New("no column " + string(c))
}

// Observation holds an arbitrary, nullable item of column data
type Observation struct {
	col   *ColumnName
	value interface{}
}

func (o Observation) ColumnName() ColumnName {
	return *o.col
}

// IsNull returns true if the value is null
func (o Observation) IsNull() bool {
	return o.value == nil
}

func (o Observation) MustInt64() int64 {
	// panic on err
	return o.value.(int64)
}

func (o Observation) MustInt() int {
	return int(o.MustInt64())
}

func (o Observation) MustString() string {
	return o.value.(string)
}

func (o Observation) Interface() interface{} {
	return o.value
}
