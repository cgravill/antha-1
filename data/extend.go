package data

import (
	"reflect"
)

/*
 * calculated columns
 */

// TODO: MustExtension for capturing errors...
// TODO figure out how to avoid recalculating state, without user having to cache
// TODO make extension series bounded and sized, if the wrapped series are
// TODO preserve sort key

// Extension is the fluent interface for adding calculated columns.
type Extension struct {
	// the new column to add
	newCol ColumnName
	// all table series
	series []*Series
}

// By allows dynamic access to observations
func (e *Extension) By(f func(r Row) interface{}, newType reflect.Type) *Table {
	// TODO: either reflectively infer newType, or assert/verify the f return type
	series := append(e.series, &Series{
		col: e.newCol,
		typ: newType,
		read: func(cache *seriesIterCache) iterator {
			return &extendRowSeries{f: f, source: e.extensionSource(cache)}
		}},
	)
	// TODO preserve sort key
	newT := NewTable(series)
	return newT
}

// extensionSource is exhausted when the underlying table is. Side effects are
// important to get table cardinality correct without requiring the extension
// column iterator to return false.
func (e *Extension) extensionSource(cache *seriesIterCache) *readRow {
	// virtual table will not be used to advance
	source := &readRow{iteratorCache: cache}
	// go get the series iterators we need from the cache
	source.fill(e.series)
	return source
}

type extendRowSeries struct {
	f      func(r Row) interface{}
	source *readRow
}

func (i *extendRowSeries) Next() bool { return true }

func (i *extendRowSeries) Value() interface{} {
	row := i.source.Value().(Row)
	v := i.f(row)
	return v
}

// On is for operations on homogeneous columns of static type
func (e *Extension) On(cols ...ColumnName) *ExtendOn {
	schema := newSchema(e.series)
	on := &ExtendOn{extension: e, combinedSeriesMeta: newExtendSeriesMeta(e.series)}
	for _, c := range cols {
		// TODO panic here, need test for this case
		on.inputs = append(on.inputs, e.series[schema.byName[c][0]])
	}
	return on
}

// Constant adds a constant column to the table
func (e *Extension) Constant(value interface{}) *Table {
	ser := &Series{
		col:  e.newCol,
		typ:  reflect.TypeOf(value),
		meta: newExtendSeriesMeta(e.series),
		read: func(cache *seriesIterCache) iterator {
			e.extensionSource(cache)
			return &constIterator{value: value}
		},
	}

	return NewTable(append(e.series, ser))
}

// NewConstantSeries is an unbounded repetition of the same value
func NewConstantSeries(col ColumnName, value interface{}) *Series {
	iter := &constIterator{value: value}
	return &Series{
		col: col,
		typ: reflect.TypeOf(value),
		read: func(_ *seriesIterCache) iterator {
			return iter
		},
	}
}

type constIterator struct {
	value interface{}
}

func (i *constIterator) Next() bool         { return true }
func (i *constIterator) Value() interface{} { return i.value }

// newExtendSeriesMeta returns nil if all the underlying series are unbounded.
// TODO extension col should be bounded if at least one underlying series is bounded
func newExtendSeriesMeta(series []*Series) *combinedSeriesMeta {
	// FIXME implement this
	return nil
}

// the nil value is unbounded (TODO does this make sense)
type combinedSeriesMeta struct{ exact, max int }

func (m *combinedSeriesMeta) IsBounded() bool      { return m != nil }
func (m *combinedSeriesMeta) IsMaterialized() bool { return false }
func (m *combinedSeriesMeta) ExactSize() int {
	if m == nil {
		return -1
	}
	return m.exact
}
func (m *combinedSeriesMeta) MaxSize() int {
	if m.IsBounded() {
		return m.max
	}
	panic("don't call MaxSize on unbounded series")
}

// ExtendOn enables extensions using specific column values as function inputs
type ExtendOn struct {
	*combinedSeriesMeta // FIXME use this value
	extension           *Extension
	inputs              []*Series
}
