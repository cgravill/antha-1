package data

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

/*
 * filter interfaces
 */

// Selection is the fluent interface for filtering rows.
type Selection struct {
	t *Table
}

// By accepts a filter function that operates on rows of the whole table.
// The filtered table contains rows where this function returns true.
func (s *Selection) By(fn MatchRow) *Table {
	return filterTable(fn, s.t)
}

// On selects columns for filtering. Note this does not panic yet, even if the
// columns do not exist.  (However subsequent calls to the returned object will
// error.)
func (s *Selection) On(cols ...ColumnName) *FilterOn {
	return &FilterOn{t: s.t, cols: cols}
}

// FilterOn filters by named columns.
// TODO(twoodwark): this is needlessly eager with respect to all underlying columns.
type FilterOn struct {
	t    *Table
	cols []ColumnName
}

// TODO
// func (o *FilterOn) Not() *FilterOn {}
// func (o *FilterOn) Null() (*Table, error) {} // all null

func (o *FilterOn) checkSchema(colsType reflect.Type, assertions ...SchemaAssertion) error {
	// eager schema check
	filterSubject, err := o.t.Project(o.cols...)
	if err != nil {
		return errors.Wrapf(err, "can't filter columns %+v", o.cols)
	}
	// assert columns assignable
	if colsType != nil {
		for _, col := range filterSubject.Schema().Columns {
			if !col.Type.AssignableTo(colsType) {
				return errors.Errorf("column %s is not assignable to type %v", col.Name, colsType)
			}
		}
	}
	for _, assrt := range assertions {
		if err := assrt(filterSubject.Schema()); err != nil {
			return err
		}
	}
	return nil
}

func (o *FilterOn) matchColIndexes(assertions ...SchemaAssertion) map[int]int {
	// which columns values do we need?
	matchColIndexes := map[int]int{}
	for i, n := range o.cols {
		for _, c := range o.t.schema.byName[n] {
			matchColIndexes[c] = i
		}
	}
	return matchColIndexes
}

// SchemaAssertion is given the schema of the table projection on which the
// filter will operate.  It should return nil if the schema is acceptable.
type SchemaAssertion func(Schema) error

// MatchRow implements a filter on entire table rows.
type MatchRow func(r Row) bool

/*
 * generic filter guts
 */

// the filtered series share an underlying iterator cache
type filterState struct {
	// matcher determines when to return the row.
	// TODO  we don't always need to read the whole Row. colVals do not need to
	// be updated for lazy columns when we already know we matched false
	// (assuming we are using column matchers and not row matchers).
	matcher  MatchRow
	source   *tableIterator
	iterNext bool
	curr     []interface{}
	index    Index
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
	row := st.source.Value().(Row)
	colVals := []interface{}{}

	for _, o := range row.Values {
		colVals = append(colVals, o.value)
	}
	// cache the column values for the underlying, in case they are expensive
	st.curr = colVals
	return st.matcher(row)
}

type filterIter struct {
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

// Value reads the cached column value
func (iter *filterIter) Value() interface{} {
	colVals := iter.commonState.curr
	return colVals[iter.colIndex]
}

// compose the matchRow filter into all the series
func filterTable(matchRow MatchRow, table *Table) *Table {
	newTable := newFromTable(table, table.sortKey...)
	wrap := func(colIndex int, wrappedSeries *Series) func(cache *seriesIterCache) iterator {
		return func(cache *seriesIterCache) iterator {
			// The first wrapper needs to construct the common state for the parent iterator,
			// noting we will be called in random order.
			var commonState *filterState
			for _, w := range newTable.series {
				if iterator, found := cache.cache[w]; found {
					commonState = iterator.(*filterIter).commonState
				}
			}
			if commonState == nil {
				commonState = &filterState{
					index:   -1,
					matcher: matchRow,
					source:  newTableIterator(table.series),
				}
			}

			return &filterIter{
				pos:         commonState.index,
				colIndex:    colIndex,
				commonState: commonState,
			}
		}
	}
	for i, wrappedSeries := range table.series {
		newTable.series[i] = &Series{
			typ:  wrappedSeries.typ,
			col:  wrappedSeries.col,
			read: wrap(i, wrappedSeries),
			meta: &filteredSeriesMeta{wrapped: wrappedSeries.meta},
		}
	}
	return newTable
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

/*
 * concrete filters
 */

// Eq retuns a function matching the selected column(s) where equal to expected
// value(s), after any required type conversion.
func Eq(expected ...interface{}) (MatchInterface, SchemaAssertion) {
	assertion := &eq{expected: expected, converted: make([]interface{}, len(expected))}
	return func(v ...interface{}) bool {
		return reflect.DeepEqual(v, assertion.converted)
	}, assertion.CheckSchema
}

type eq struct {
	expected, converted []interface{}
}

// TODO Eq specialization methods to more efficiently filter known scalar types (?)

// CheckSchema converts expected values, as a side effect
func (w *eq) CheckSchema(schema Schema) error {
	if schema.Size() != len(w.expected) {
		return fmt.Errorf("Eq: %d column(s), to equal %d value(s) %+v", schema.Size(), len(w.expected), w.expected)
	}
	for i, c := range schema.Columns {
		// convert to the column type
		val := reflect.ValueOf(w.expected[i])
		if !val.Type().ConvertibleTo(c.Type) {
			return fmt.Errorf("Eq: inconvertible type for %s: %+v to %+v", c.Name, val.Type(), c.Type)
		}
		w.converted[i] = val.Convert(c.Type).Interface()
	}
	w.expected = nil
	return nil
}
