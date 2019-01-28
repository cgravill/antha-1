package data

import (
	"github.com/apache/arrow/go/arrow/array"
	"reflect"
)

/*
 * utility for wrapping error functions
 */

func handle(err error) {
	if err != nil {
		panic(err)
	}
}

// MustCreate panics on any error when creating tables/series
type MustCreate struct{}

func Must() MustCreate {
	return MustCreate{}
}

func (m MustCreate) NewSliceSeries(col ColumnName, values interface{}) *Series {
	ser, err := NewSliceSeries(col, values)
	handle(err)
	return ser
}

func (m MustCreate) NewArrowSeries(col ColumnName, values array.Interface) *Series {
	ser, err := NewArrowSeries(col, values)
	handle(err)
	return ser
}

func (m MustCreate) NewArrowSeriesFromSlice(col ColumnName, values interface{}, mask []bool) *Series {
	ser, err := NewArrowSeriesFromSlice(col, values, mask)
	handle(err)
	return ser
}

func (m MustCreate) NewTableFromStructs(structs interface{}) *Table {
	t, err := NewTableFromStructs(structs)
	handle(err)
	return t
}

type MustSeries struct {
	s *Series
}

func (s *Series) Must() MustSeries {
	return MustSeries{s: s}
}

func (m MustSeries) Cache() *Series {
	s, err := m.s.Cache()
	handle(err)
	return s
}

type MustTable struct {
	t *Table
}

func (t *Table) Must() MustTable {
	return MustTable{t: t}
}

func (m MustTable) Cache() *Table {
	t, err := m.t.Cache()
	handle(err)
	return t
}

func (m MustTable) Project(columns ...ColumnName) *Table {
	t, err := m.t.Project(columns...)
	handle(err)
	return t
}

func (m MustTable) Convert(col ColumnName, typ reflect.Type) *Table {
	t, err := m.t.Convert(col, typ)
	handle(err)
	return t
}

func (m MustTable) Filter() *MustSelection {
	return &MustSelection{m.t.Filter()}
}

type MustSelection struct {
	*Selection
}

func (s *MustSelection) On(cols ...ColumnName) *MustFilterOn {
	return &MustFilterOn{s.Selection.On(cols...)}
}

type MustFilterOn struct {
	*FilterOn
}

func (o *MustFilterOn) Interface(m MatchInterface, assertions ...SchemaAssertion) *Table {
	t, err := o.FilterOn.Interface(m, assertions...)
	handle(err)
	return t
}
