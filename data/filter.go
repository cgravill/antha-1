package data

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

/*
 * filter interfaces
 */

// Matcher returns true for column values that match
type Matcher func(...interface{}) bool

// FilterSpec implements a filter operation
type FilterSpec interface {
	Columns() []ColumnName
	// MatchFor should return error if the filter is invalid eg. requires a different datatype
	// the required columnnames are guaranteed to be present in this schema
	MatchFor(Schema) (Matcher, error)
	// TODO bridge interfaces for more efficient backend filter ops
}

/*
 * concrete filters
 */
type eq struct {
	Col ColumnName
	Val interface{}
}

// Eq matches Rows that have the given column equal to val
func Eq(col ColumnName, val interface{}) FilterSpec {
	return eq{Col: col, Val: val}
}

// TODO Eq specialization methods to more efficiently filter known scalar types (?)

func (w eq) Columns() []ColumnName {
	return []ColumnName{w.Col}
}

func (w eq) MatchFor(schema Schema) (Matcher, error) {
	c, _ := schema.Col(w.Col)
	// if possible, cast/convert to the matched type
	val := reflect.ValueOf(w.Val)
	if !val.Type().ConvertibleTo(c.Type) {
		return nil, fmt.Errorf("Incompatible types: %+v and %+v", val.Type(), c.Type)
	}
	valConverted := val.Convert(c.Type).Interface()
	matchValues := func(v ...interface{}) bool {
		return v[0] == valConverted
	}

	return matchValues, nil
}

/*
 * generic filter guts
 */
type filterState struct {
	matchColIndexes map[int]bool
	matcher         Matcher
	source          *tableIterator
	iterNext        bool
	curr            []interface{}
	index           Index
}

// the filtered series share an underlying iterator cache
func newFilterState(matcher Matcher, filterSpec FilterSpec, underlying []*Series) *filterState {
	s := &filterState{
		index:           -1,
		matcher:         matcher,
		matchColIndexes: map[int]bool{},
		source:          newTableIterator(underlying),
	}
	schema := newSchema(underlying)
	for _, n := range filterSpec.Columns() {
		for _, c := range schema.byName[n] {
			s.matchColIndexes[c] = true
		}
	}

	return s
}

func (st *filterState) advance() {
	for st.source.Next() {
		st.index++
		if st.isMatch() {
			st.iterNext = true
			return
		}
	}
	st.iterNext = false
}
func (st *filterState) isMatch() bool {
	// TODO save some garbage here by indexing columns by position and only
	// getting series values needed to satisfy the filter (ie avoid copying to
	// a Row!)
	row := st.source.Value().(Row)
	colVals := []interface{}{}
	matchVals := []interface{}{}
	for i, n := range st.source.cols {
		o, _ := row.Observation(*n)
		colVals = append(colVals, o.value)
		if st.matchColIndexes[i] {
			matchVals = append(matchVals, o.value)
		}
	}
	st.curr = colVals
	return st.matcher(matchVals...)
}

type filterIter struct {
	wrapped     iterator // series iterator
	commonState *filterState
	pos         Index
	colIndex    int
}

// Next must not return until the source has been advanced to
// a true filter state, or has been exhausted.
func (iter *filterIter) Next() bool {
	// see if we need to discard the current shared state
	retain := iter.pos != iter.commonState.index
	if !retain {
		iter.commonState.advance()
		iter.pos = iter.commonState.index
	}
	return iter.commonState.iterNext
}

func (iter *filterIter) Value() interface{} {
	colVals := iter.commonState.curr
	return colVals[iter.colIndex]
}

func lazyFilterTable(filterSpec FilterSpec, table *Table) (*Table, error) {
	// eager schema check
	filterSubject, err := table.Project(filterSpec.Columns()...)
	if err != nil {
		return nil, errors.Wrapf(err, "can't filter columns %+v", filterSpec.Columns())
	}
	matcher, err := filterSpec.MatchFor(filterSubject.Schema())
	if err != nil {
		panic(errors.Wrapf(err, "can't filter %+v with %+v", table, filterSpec))
	}
	// compose the filter into all the series
	wrappers := make([]*Series, len(table.series))
	wrap := func(colIndex int, wrappedSeries *Series) func(cache *seriesIterCache) iterator {
		return func(cache *seriesIterCache) iterator {
			// The first wrapper needs to construct the common state for the parent iterator,
			// noting we will be called in random order.
			var commonState *filterState
			for _, w := range wrappers {
				if iterator, found := cache.cache[w]; found {
					commonState = iterator.(*filterIter).commonState
				}
			}
			if commonState == nil {
				commonState = newFilterState(matcher, filterSpec, table.series)
			}

			return &filterIter{
				pos:         commonState.index,
				colIndex:    colIndex,
				commonState: commonState,
				wrapped:     commonState.source.iteratorCache.Ensure(wrappedSeries),
			}
		}
	}
	for i, wrappedSeries := range table.series {
		wrappers[i] = &Series{
			typ:  wrappedSeries.typ,
			col:  wrappedSeries.col,
			read: wrap(i, wrappedSeries),
			meta: &filteredSeriesMeta{wrapped: wrappedSeries.meta},
		}
	}
	return NewTable(wrappers), nil
}

// filtered series metadata
type filteredSeriesMeta struct {
	wrapped SeriesMeta
}

func (m *filteredSeriesMeta) IsBounded() bool      { return m.wrapped.IsBounded() }
func (m *filteredSeriesMeta) IsMaterialized() bool { return false }

func (m *filteredSeriesMeta) ExactSize() int {
	return -1
}

func (m *filteredSeriesMeta) MaxSize() int {
	if m.IsBounded() {
		return m.wrapped.(Bounded).MaxSize()
	} else {
		return -1
	}
}
