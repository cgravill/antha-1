package data

import (
	"github.com/apache/arrow/go/arrow/array"
)

/*
 * utility for wrapping error functions
 */

type MustCreate struct{}

// Must asserts no error on the objects it creates
func Must() MustCreate {
	return MustCreate{}
}

func (m MustCreate) handle(err error) {
	if err != nil {
		panic(err)
	}
}

func (m MustCreate) NewSliceSeries(col ColumnName, values interface{}) *Series {
	ser, err := NewSliceSeries(col, values)
	m.handle(err)
	return ser
}

func (m MustCreate) NewArrowSeries(col ColumnName, values array.Interface) *Series {
	ser, err := NewArrowSeries(col, values)
	m.handle(err)
	return ser
}

func (m MustCreate) NewArrowSeriesFromSlice(col ColumnName, values interface{}, mask []bool) *Series {
	ser, err := NewArrowSeriesFromSlice(col, values, mask)
	m.handle(err)
	return ser
}

func (m MustCreate) NewTableFromStructs(structs interface{}) *Table {
	t, err := NewTableFromStructs(structs)
	m.handle(err)
	return t
}

type MustSeries struct {
	s *Series
}

func (m MustSeries) handle(err error) {
	if err != nil {
		panic(err)
	}
}

func (s *Series) Must() MustSeries {
	return MustSeries{s: s}
}

func (m MustSeries) Cache() *Series {
	s, err := m.s.Cache()
	m.handle(err)
	return s
}

type MustTable struct {
	t *Table
}

func (m MustTable) handle(err error) {
	if err != nil {
		panic(err)
	}
}

func (t *Table) Must() MustTable {
	return MustTable{t: t}
}

func (m MustTable) Cache() *Table {
	t, err := m.t.Cache()
	m.handle(err)
	return t
}
