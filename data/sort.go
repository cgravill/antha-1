package data

import (
	"sort"
)

// For now, settled on a very simple implementation of table sorting - via materialization into a slice of rows
func sortTable(t *Table, predicate func(r1 *Row, r2 *Row) bool) (*Table, error) {
	// a Row-wide representation of a table
	rows := t.ToRows()

	// sorting by a user-defined predicate
	sort.SliceStable(rows.Data, func(i, j int) bool {
		return predicate(&rows.Data[i], &rows.Data[j])
	})

	// converting into Arrow-based series
	series := make([]*Series, len(rows.Schema.Columns))
	for i := range series {
		var err error
		series[i], err = NewArrowSeriesFromRows(&rows, rows.Schema.Columns[i].Name)
		if err != nil {
			return nil, err
		}
	}

	return NewTable(series), nil
}

// TODO: predefined ascending/descending sorting functions (generated)
