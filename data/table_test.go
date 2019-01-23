package data

import (
	"fmt"
	"reflect"
	"testing"
	// TODO "github.com/stretchr/testify/assert"
)

func TestEquals(t *testing.T) {
	testEquals(t, nativeSeries)
	testEquals(t, arrowSeries)
}

func TestEqualsComplexType(t *testing.T) {
	assertEqual(t, NewTable([]*Series{
		newSeries(nativeSeries, "y", []int32{}),
		newSeries(nativeSeries, "x", [][]string{}),
	}), NewTable([]*Series{
		newSeries(nativeSeries, "y", []int32{}),
		newSeries(nativeSeries, "x", [][]string{}),
	}), "not equal")

}

func testEquals(t *testing.T, typ seriesType) {
	tab := NewTable([]*Series{
		newSeries(typ, "measure", []int64{1, 1000}),
		newSeries(typ, "label", []string{"abcdef", "abcd"}),
	})
	assertEqual(t, tab, tab, "not self equal")

	tab2 := NewTable([]*Series{
		newSeries(typ, "measure", []int64{1, 1000}),
	})
	assertEqual(t, tab2, tab.Project("measure"), "not equal by value")

	if tab2.Equal(tab.Project("label")) {
		t.Error("equal with mismatched schema")
	}

	if tab2.Equal(tab2.Filter(Eq("measure", 1000))) {
		t.Error("equal with mismatched data")
	}
}

func assertEqual(t *testing.T, expected, actual *Table, msg string) {
	if !actual.Equal(expected) {
		t.Error(msg)
		t.Log("actual", actual.Head(20).ToRows())
	}
}

func TestSlice(t *testing.T) {
	testSlice(t, nativeSeries)
	testSlice(t, arrowSeries)
}

func testSlice(t *testing.T, typ seriesType) {
	a := NewTable([]*Series{
		newSeries(typ, "a", []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}),
	})
	assertEqual(t, a, a.Slice(0, 100), "slice all")

	slice00 := a.Slice(1, 1)
	assertEqual(t, NewTable([]*Series{
		newSeries(typ, "a", []int64{}),
	}), slice00, "slice00")

	slice04 := a.Head(4)
	assertEqual(t, NewTable([]*Series{
		newSeries(typ, "a", []int64{1, 2, 3, 4}),
	}), slice04, "slice04")

	slice910 := a.Slice(9, 10)
	assertEqual(t, NewTable([]*Series{
		newSeries(typ, "a", []int64{10}),
	}), slice910, "slice910")
}

func TestExtend(t *testing.T) {
	testExtend(t, nativeSeries)
	testExtend(t, arrowSeries)
}

func testExtend(t *testing.T, typ seriesType) {
	a := NewTable([]*Series{
		newSeries(typ, "a", []int64{1, 2, 3}),
	})
	extended := a.Extend("e").By(func(r Row) interface{} {
		a, _ := r.Observation("a")
		return float64(a.MustInt64()) / 2.0
	},
		reflect.TypeOf(float64(0)))
	assertEqual(t, NewTable([]*Series{
		newSeries(typ, "e", []float64{0.5, 1.0, 1.5}),
	}), extended.Project("e"), "extend")

	floats := NewTable([]*Series{
		newSeries(typ, "floats", []float64{1, 2, 3}),
	})
	extendedStatic := floats.
		Extend("e_static").
		On("floats").
		Float64(func(v ...float64) float64 {
			return v[0] * 2.0
		})

	assertEqual(t, NewTable([]*Series{
		Must().NewSliceSeries("e_static", []float64{2, 4, 6}),
	}), extendedStatic.Project("e_static"), "extend static")

	// you don't actually need to set any inputs!
	// note that an impure extension is bad practice in general.
	i := int64(0)
	extendedStaticNullary := floats.
		Extend("generator").
		On().
		Int64(func(_ ...int64) int64 {
			i++
			return i * 10
		})

	assertEqual(t, NewTable([]*Series{
		Must().NewSliceSeries("generator", []int64{10, 20, 30}),
	}), extendedStaticNullary.Project("generator"), "generator")

	extendedConst := floats.
		Extend("constant").
		Constant(float64(8))
	assertEqual(t, NewTable([]*Series{
		Must().NewSliceSeries("constant", []float64{8, 8, 8}),
	}), extendedConst.Project("constant"), "extend const")
}

func TestConstantColumn(t *testing.T) {
	tab := NewTable([]*Series{NewConstantSeries("a", 1)}).Head(2)
	assertEqual(t, NewTable([]*Series{
		Must().NewSliceSeries("a", []int{1, 1}),
	}), tab, "const")
}

func TestRename(t *testing.T) {
	tab := NewTable([]*Series{NewConstantSeries("a", 1)}).Rename("a", "x").Head(2)
	assertEqual(t, NewTable([]*Series{
		Must().NewSliceSeries("x", []int{1, 1}),
	}), tab, "renamed")
}

func TestFilterEq(t *testing.T) {
	testFilterEq(t, nativeSeries)
	testFilterEq(t, arrowSeries)
}

func testFilterEq(t *testing.T, typ seriesType) {
	a := NewTable([]*Series{
		newSeries(typ, "a", []int64{1, 2, 3}),
	})
	filtered := a.Filter(Eq("a", 2))
	assertEqual(t, filtered, a.Slice(1, 2), "filter")
}

func TestSize(t *testing.T) {
	testSize(t, nativeSeries)
	testSize(t, arrowSeries)
}

func testSize(t *testing.T, typ seriesType) {
	empty := NewTable([]*Series{})
	if empty.Size() != 0 {
		t.Errorf("should be empty. %d", empty.Size())
	}
	a := NewTable([]*Series{
		newSeries(typ, "a", []int64{1, 2, 3}),
	})
	if a.Size() != 3 {
		t.Errorf("size? %d", a.Size())
	}
	// a filter is of unbounded size
	filtered := a.Filter(Eq("a", 1))
	if filtered.Size() != -1 {
		t.Errorf("filtered.Size()? %d", filtered.Size())
	}
	// a slice is of bounded size as long as its dependencies are
	slice1 := filtered.Head(1)
	if slice1.Size() != -1 {
		t.Errorf(" slice1.Size()? %d", slice1.Size())
	}
	if a.Head(0).Size() != 0 {
		t.Errorf("a.Head(0).Size()? %d", a.Head(0).Size())
	}
	slice2 := a.Slice(1, 4)
	if slice2.Size() != 2 {
		t.Errorf("slice2.Size()? %d", slice2.Size())
	}
}

func TestCache(t *testing.T) {
	testCache(t, nativeSeries)
	testCache(t, arrowSeries)
}

// TODO: .Cache must work on arbitrary series types
func testCache(t *testing.T, typ seriesType) {
	// a materialized table of 3 elements
	a := NewTable([]*Series{
		newSeries(typ, "a", []int64{1, 2, 3}),
	})

	// a lazy table - after filtration
	filtered := a.Filter(Eq("a", 1))

	// a materialized copy
	filteredCached, err := filtered.Cache()
	if err != nil {
		t.Errorf("cache failed: %s", err)
	}

	// check the cached table has the same content
	assertEqual(t, filtered, filteredCached, "copy")
	// check the copy size
	if filteredCached.Size() != 1 {
		t.Errorf("filteredCached.Size()? %d", filteredCached.Size())
	}
}

func TestSort(t *testing.T) {
	// an input table - sorted by id
	table := NewTable([]*Series{
		Must().NewArrowSeriesFromSlice("id", []int64{1, 2, 3, 4, 5}, nil),
		Must().NewArrowSeriesFromSlice("int64_measure", []int64{50, 20, 20, 20, 10}, nil),
		Must().NewArrowSeriesFromSlice("float64_nullable_measure", []float64{1., -1., 2., 2., 5.}, []bool{true, false, true, true, true}),
	})

	// sorting the table by two other columns
	sorted, err := table.Sort([]ColumnKey{
		ColumnKey{Column: "int64_measure", Asc: true},
		ColumnKey{Column: "float64_nullable_measure", Asc: false},
	})
	if err != nil {
		t.Errorf("sort failed: %s", err)
	}

	// reference sorted table
	sortedReference := NewTable([]*Series{
		Must().NewArrowSeriesFromSlice("id", []int64{5, 3, 4, 2, 1}, nil), // 1 and 5 should swap; 3 and 4 should remain in the same order (since sorting is stable)
		Must().NewArrowSeriesFromSlice("int64_measure", []int64{10, 20, 20, 20, 50}, nil),
		Must().NewArrowSeriesFromSlice("float64_nullable_measure", []float64{5., 2., 2., -1., 1.}, []bool{true, true, true, false, true}),
	})

	assertEqual(t, sortedReference, sorted, "sort")
}

func TestSortByFunc(t *testing.T) {
	testSortByFunc(t, nativeSeries)
	testSortByFunc(t, arrowSeries)
}

func testSortByFunc(t *testing.T, typ seriesType) {
	// an unsorted table
	table := NewTable([]*Series{
		newSeries(typ, "id", []int64{2, 1, 3}),
		newSeries(typ, "str", []string{"2", "1", "3"}),
	})

	// a table sorted by id
	sorted, err := table.SortByFunc(func(r1 *Row, r2 *Row) bool {
		return r1.Values[0].MustInt64() < r2.Values[0].MustInt64()
	})
	if err != nil {
		t.Errorf("sort by func failed: %s", err)
	}

	// sorted table reference value
	sortedReference := NewTable([]*Series{
		newSeries(typ, "id", []int64{1, 2, 3}),
		newSeries(typ, "str", []string{"1", "2", "3"}),
	})

	// check the sorted table is equal to the reference table
	assertEqual(t, sortedReference, sorted, "sort by func")
}

type seriesType int

const (
	nativeSeries seriesType = iota
	arrowSeries
)

func newSeries(typ seriesType, col ColumnName, values interface{}) *Series {
	switch typ {
	case nativeSeries:
		return Must().NewSliceSeries(col, values)
	case arrowSeries:
		return Must().NewArrowSeriesFromSlice(col, values, nil)
	default:
		panic(fmt.Errorf("Unknown series type: %d", typ))
	}
}
