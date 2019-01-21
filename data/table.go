package data

//go:generate python gen.py
import (
	"math"
	"reflect"

	"github.com/pkg/errors"
)

// Lazy data sets

// Table is an immutable container of Series
// It can optionally be keyed.
type Table struct {
	series []*Series
	// this must return Row
	read func([]*Series) *tableIterator
}

// NewTable gives lowlevel access.  TODO semantics if jagged columns, duplicates etc
func NewTable(series []*Series) *Table {
	return &Table{
		series: series,
		read:   newTableIterator,
	}
}

// Schema returns the type information for the Table
func (t *Table) Schema() Schema {
	return newSchema(t.series)
}

// TODO do we need this
func (t *Table) Series() []*Series {
	return nil
}

func (t *Table) seriesMap() map[ColumnName][]*Series {

	schema := t.Schema()
	byName := map[ColumnName][]*Series{}
	for n, is := range schema.byName {
		for _, i := range is {
			byName[n] = append(byName[n], t.series[i])
		}
	}
	return byName
}

// TODO do we need this
// func (t *Table) ColumnNames() []ColumnName {
// 	return nil
// }

// IterAll iterates over the entire table, no buffer.
// Use when ranging over all rows is required.
func (t *Table) IterAll() <-chan Row {
	rows, _ := t.Iter()
	return rows
}

// Iter iterates over the table, no buffer.
// call done() to release resources after a partial read.
func (t *Table) Iter() (rows <-chan Row, done func()) {
	channel := make(chan Row)
	iter := t.read(t.series)
	control := make(chan struct{}, 1)
	done = func() {
		control <- struct{}{}
	}
	go func() {
		defer close(channel)
		for iter.Next() {
			rowRaw := iter.Value()
			row := rowRaw.(Row)
			select {
			case <-control:
				return
			case channel <- row:
				// do nothing
			}
		}
	}()
	return channel, done
}

// ToRows materializes data: may be very expensive
func (t *Table) ToRows() Rows {
	rr := Rows{Schema: t.Schema()}
	for r := range t.IterAll() {
		rr.Data = append(rr.Data, r)
	}
	return rr
}

// Slice is a lazy subset of records between the start index and the end (exclusive)
// unlike go slices, if the end index is out of range then fewer records are returned
// rather than receiving an error
func (t *Table) Slice(start, end Index) *Table {
	newSeries := make([]*Series, len(t.series))
	for i, ser := range t.series {
		m := newSeriesSlice(ser, start, end)
		newSeries[i] = &Series{
			typ:  ser.typ,
			col:  ser.col,
			read: m.read,
			meta: m,
		}
	}
	return NewTable(newSeries)
}

// Head is a lazy subset of the first count records (but may return fewer)
func (t *Table) Head(count int) *Table {
	return t.Slice(0, Index(count))
}

// Key returns the key columns
func (t *Table) Key() Key {
	return nil
}

// WithKey sets the sort key (but does not sort).
func (t *Table) WithKey(key Key) *Table {
	return nil
}

// Sort produces a sorted Table using the Key.
// TODO inplace optimization?
//func (t *Table) Sort(asc ...bool) *Table {
//
//}

// Sort sorts a table by a user-defined function
// In order not to run out of resources, it is recommended to call sortTable only after removing unnecessary columns with .Project(...)
// TODO: decide whether we prefer this more general sorting or specifying table keys (and sorting only by them) instead?
func (t *Table) Sort(f func(r1 *Row, r2 *Row) bool) (*Table, error) {
	return sortTable(t, f)
}

// Equal is true if the other table has the same schema (in the same order)
// and exactly equal series values
func (t *Table) Equal(other *Table) bool {
	if t == other {
		return true
	}
	schema1 := t.Schema()
	schema2 := other.Schema()
	if !schema1.Equal(schema2) {
		return false
	}
	// TODO compare tables' key, known bounded length, and sortedness

	// TODO if table series are identical we can shortcut the iteration
	iter1, done1 := t.Iter()
	iter2, done2 := other.Iter()
	defer done1()
	defer done2()
	for {
		r1, more1 := <-iter1
		r2, more2 := <-iter2
		if more1 != more2 || !reflect.DeepEqual(r1.Values, r2.Values) {
			return false
		}
		if !more1 {
			break
		}
	}

	return true
	// TODO since we are iterating over possibly identical series we might optimize by sharing the iterator cache
}

// Size returns -1 if unknwon (because unbounded or lazy)
func (t *Table) Size() int {
	if len(t.series) == 0 {
		return 0
	}
	max := math.MaxInt64
	exact := math.MaxInt64
	for _, ser := range t.series {
		if b, ok := ser.meta.(Bounded); ok {
			sMax := b.MaxSize()
			if sMax == 0 {
				return 0
			} else if sMax < max {
				max = sMax
			}
			sX := b.ExactSize()
			if sX < exact {
				exact = sX
			}
		} else {
			// unbounded
			exact = -1
		}
	}
	return exact
}

// Cache converts a lazy table to one that is fully materialized
func (t *Table) Copy() (*Table, error) {
	seriesCopies := make([]*Series, len(t.series))
	for i, series := range t.series {
		if seriesCopy, err := series.Copy(); err != nil {
			return nil, err
		} else {
			seriesCopies[i] = seriesCopy
		}
	}
	return NewTable(seriesCopies), nil
}

// Cache converts a lazy table to one that is fully materialized
func (t *Table) Cache() (*Table, error) {
	for _, series := range t.series {
		// TODO is it OK that cached series are shared with other instances?
		if _, err := series.Cache(); err != nil {
			return nil, err
		}
	}
	return t, nil
}

// DropNullColumns filters out columns with all/any row null
// TODO
func (t *Table) DropNullColumns(all bool) *Table {
	return nil
}

// DropNull filters out rows with all/any col null
// TODO
func (t *Table) DropNull(all bool) *Table {
	return nil
}

// Project reorders and/or takes a subset of columns.
// On duplicate columns, only the first so named is taken.
func (t *Table) Project(columns ...ColumnName) *Table {
	s := make([]*Series, len(columns))
	byName := t.seriesMap()
	for i, n := range columns {
		if sers, found := byName[n]; !found {
			panic(errors.Errorf("cannot project %v, no such column '%s'", t.Schema(), n))
		} else {
			s[i] = sers[0] //!
		}
	}
	return NewTable(s)
}

// ProjectAllBut discards the named columns, which may not exist in the schema
func (t *Table) ProjectAllBut(columns ...ColumnName) *Table {
	byName := map[ColumnName]struct{}{}
	for _, n := range columns {
		byName[n] = struct{}{}
	}
	s := []*Series{}
	for _, ser := range t.series {
		if _, found := byName[ser.col]; !found {
			s = append(s, ser)
		}
	}
	return NewTable(s)
}

// Filter selects some records lazily
func (t *Table) Filter(f FilterSpec) *Table {
	return lazyFilterTable(f, t)
}

// Join is a natural join on sorted tables with the same Key
// TODO dedup series (?)
func (t *Table) Join(other Joinable) *Table {
	return nil
}

// Extend adds a column by applying a function
func (t *Table) Extend(newCol ColumnName) *Extension {
	series := make([]*Series, len(t.series))
	copy(series, t.series)
	return &Extension{newCol: newCol, series: series}
}

// Copy gives a new table, optionally with duplicate Series data
// TODO semantics
// func (t *Table) Copy(deep bool) *Table {
// 	return nil
// }

var _ Joinable = (*Table)(nil)

type Joinable interface {
	//?
	Key() Key
	Series() []*Series
	// TODO bridge interfaces for more efficient backend join ops
}
