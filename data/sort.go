package data

import (
	"sort"
	"strings"
)

func sortTableByKey(t *Table, key Key) (*Table, error) {
	// short path - in case the table is already sorted by the same (or more specialized) key
	if tableKey := t.Schema().Key; tableKey != nil && tableKey.HasPrefix(key) {
		//TODO: should we replace the table key with key?
		return t, nil
	}

	// creating sorting function
	sortFunc, err := createSortFunc(t.schema, key)
	if err != nil {
		return nil, err
	}

	// sorting with it
	sorted, err := sortTableByFunc(t, sortFunc)
	if err != nil {
		return nil, err
	}

	// supplying the sorted table with a key
	return sorted.withKey(key), nil
}

// creates a sort function for comparing rows which satisfy a given schema by a given key
func createSortFunc(schema Schema, key Key) (SortFunc, error) {
	// preparation: creating sorting functions for individual columns
	columnSortFuncs := make([]sortFuncExt, len(key))
	for i := range key {
		if columnSortFunc, err := createColumnSortFunc(schema, key[i]); err != nil {
			return nil, err
		} else {
			columnSortFuncs[i] = columnSortFunc
		}
	}

	// creating a compound sorting function
	return func(r1 *Row, r2 *Row) bool {
		for _, columnSortFunc := range columnSortFuncs {
			result := columnSortFunc(r1, r2)
			if result != 0 {
				return result < 0
			}
		}
		return false
	}, nil
}

// For now, settled on a very simple implementation of sorting by a function - via materialization into a slice of rows
func sortTableByFunc(t *Table, predicate func(r1 *Row, r2 *Row) bool) (*Table, error) {
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

// an extended sort function - returns int (needed for making compound sort functions)
type sortFuncExt func(r1 *Row, r2 *Row) int

// excluded from code generation because bools do not support comparison by <
func compareBool(val1, val2 bool) int {
	switch {
	case !val1 && val2:
		return -1
	case val1 && !val2:
		return 1
	default:
		return 0
	}
}

// excluded from code generation because using strings.Compare is more efficient
func compareString(val1, val2 string) int {
	return strings.Compare(val1, val2)
}
