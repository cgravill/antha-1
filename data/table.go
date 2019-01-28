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
	series  []*Series
	schema  Schema
	sortKey Key
	// this must return Row
	read func([]*Series) *tableIterator
}

// NewTable gives lowlevel access.
// TODO if given bounded columns of known different sizes, then return error!
func NewTable(series []*Series) *Table {
	return &Table{
		series: series,
		schema: newSchema(series),
		read:   newTableIterator,
	}
}

// newFromTable creates a new table pointing with the same series, and sets the sort key
func newFromTable(t *Table, key ...ColumnKey) *Table {
	s := make([]*Series, len(t.series))
	copy(s, t.series)
	return &Table{
		series:  s,
		schema:  t.schema,
		sortKey: key,
		read:    t.read,
	}
}

// Schema returns the type information for the Table
func (t *Table) Schema() Schema {
	return t.schema
}

// SeriesByName returns series by its name
func (t *Table) SeriesByName(col ColumnName) (*Series, error) {
	if index, err := t.schema.ColIndex(col); err == nil {
		return t.series[index], nil
	} else {
		return nil, err
	}
}

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
	newTable := newFromTable(t, t.sortKey...)
	for i, ser := range t.series {
		m := newSeriesSlice(ser, start, end)
		newTable.series[i] = &Series{
			typ:  ser.typ,
			col:  ser.col,
			read: m.read,
			meta: m,
		}
	}
	return newTable
}

// Head is a lazy subset of the first count records (but may return fewer)
func (t *Table) Head(count int) *Table {
	return t.Slice(0, Index(count))
}

// Sort produces a Table sorted by the columns defined by the Key.
// TODO inplace optimization?
func (t *Table) Sort(key Key) (*Table, error) {
	return sortTableByKey(t, key)
}

type SortFunc func(r1 *Row, r2 *Row) bool

// SortByFunc sorts a table by an arbitrary user-defined function.
// In order not to run out of resources, it is recommended to call SortByFunc only after removing unnecessary columns with .Project(...)
func (t *Table) SortByFunc(f SortFunc) (*Table, error) {
	return sortTableByFunc(t, f)
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
	// TODO compare tables' key/sortedness (?), known bounded length

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

// Size returns -1 if unknown (because unbounded or lazy)
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
func (t *Table) Cache() (*Table, error) {
	newTable := newFromTable(t, t.sortKey...)
	for i, series := range t.series {
		var err error
		if newTable.series[i], err = series.Cache(); err != nil {
			return nil, err
		}
	}
	return newTable, nil
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

// Project reorders and/or takes a subset of columns. On duplicate columns, only
// the first so named is taken. Returns error, and nil table, if any column is
// missing
func (t *Table) Project(columns ...ColumnName) (*Table, error) {
	s := make([]*Series, len(columns))
	for i, columnName := range columns {
		if series, err := t.SeriesByName(columnName); err != nil {
			return nil, errors.Wrapf(err, "when projecting %v", t.Schema())
		} else {
			s[i] = series
		}
	}
	// TODO rearrange key
	return NewTable(s), nil
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
	return NewTable(s) // TODO set key to subkey
}

// Rename updates all columns of the old name to the new name.
// Does nothing if none match in the schema.
func (t *Table) Rename(old, new ColumnName) *Table {
	s := make([]*Series, len(t.series))
	for i, ser := range t.series {
		if ser.col == old {
			ser = &Series{col: new,
				typ:  ser.typ,
				meta: ser.meta,
				read: ser.read,
			}
		}
		s[i] = ser
	}
	// TODO rename key column
	return NewTable(s)
}

// Convert lazily converts all columns of the given name to the assigned type.
// Returns non-nil error (and nil Table) if any column is not convertible. Note
// that if no column name matches, the same table is returned.
func (t *Table) Convert(col ColumnName, typ reflect.Type) (*Table, error) {
	newT := newFromTable(t, t.sortKey...)
	converted := false
	conv := &conversion{newType: typ, Table: t}
	for i, ser := range t.series {
		var err error
		if ser.col == col && !typesEqual(ser.typ, typ) {
			newT.series[i], err = conv.convert(ser)
			converted = true
			if err != nil {
				return nil, errors.Wrapf(err, "cannot convert column %d (%q)", i, col)
			}
		}
	}
	if !converted {
		return t, nil
	}
	// types but not sort key are different.  TODO: unless natural order is different for the converted type?
	newT.schema = newSchema(newT.series)
	return newT, nil
}

// Filter selects some records lazily
func (t *Table) Filter() *Selection {
	return &Selection{t}
}

// Extend adds a column by applying a function
func (t *Table) Extend(newCol ColumnName) *Extension {
	series := make([]*Series, len(t.series))
	copy(series, t.series)
	return &Extension{newCol: newCol, series: series}
}
