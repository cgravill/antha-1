package main

import (
	"fmt"
	"reflect"

	"github.com/antha-lang/antha/data"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
)

func SliceSeries() *data.Table {
	// Creating a bounded table from scalar slices.  This means reflection but
	// no copying; nullable series are not supported.

	column1 := data.Must().NewSliceSeries("measure", []int64{10, 10, 30, 0, 5})
	column2 := data.Must().NewSliceSeries("label", []string{"", "aa", "xx", "aa", ""})

	return data.NewTable([]*data.Series{column1, column2})
}

func ArrowSeries() *data.Table {
	// This example is the same as the previous one, but:
	// 1) the Series are Arrow-based (instead of slice-based ones)
	// 2) nulls are allowed

	// create the arrow series from slices and validity masks
	column1 := data.Must().NewArrowSeriesFromSlice("measure", []int64{10, 10, 30, 0, 5}, []bool{true, true, true, false, true})
	column2 := data.Must().NewArrowSeriesFromSlice("label", []string{"", "aa", "xx", "aa", ""}, nil)

	// create the Arrow series manually using builder
	builder := array.NewFloat64Builder(memory.DefaultAllocator)
	builder.Append(5.1)
	builder.AppendValues([]float64{2, 3, 0}, []bool{true, true, false})
	builder.AppendNull()
	arrowArray := builder.NewFloat64Array()

	column3 := data.NewArrowSeriesFloat64("float_measure", arrowArray)

	return data.NewTable([]*data.Series{column1, column2, column3})
}

func Example(tab *data.Table) {
	// just print Table as a whole.
	fmt.Println("input\n", tab.ToRows())

	// iterate over the entire Table.
	for record := range tab.IterAll() {
		// s := &struct{ x int64}
		// err := record.ToStruct(&s)
		// s.x
		m, _ := record.Observation("measure")
		if m.IsNull() {
			fmt.Println("measure=null at index", record.Index)
		} else {
			fmt.Println("measure=", m.MustInt64())
		}
	}

	// subset of rows
	fmt.Println("tab.Slice(2,4)\n", tab.Slice(2, 4).ToRows())

	// produce a new Table by filtering
	smallerTab := tab.Filter(data.Eq("label", "aa"))
	fmt.Println("after filter\n", smallerTab.ToRows())

	mult := func(r data.Row) interface{} {
		m, _ := r.Observation("measure")
		if m.IsNull() {
			return nil
		} else {
			return float64(m.MustInt64()) * float64(2.5)
		}
	}
	extended := tab.
		Extend("multiplied").By(mult, reflect.TypeOf(float64(0)))
	fmt.Println("extended and filtered\n", extended.Filter(data.Eq("multiplied", 25)).ToRows())

	// TODO: use static extend in the example - requires converting "measure" to homogeneous float type
	projected := tab.
		Extend("multiplied").By(mult, reflect.TypeOf(float64(0))).
		Project("label", "multiplied")
	fmt.Println("extended and projected\n", projected.ToRows())

	alternateProjected := extended.ProjectAllBut("measure")
	fmt.Printf("alternateProjected.Equal(projected): %v\n", alternateProjected.Equal(projected))
}

func main() {
	fmt.Println("slices")
	Example(SliceSeries())
	fmt.Println("arrow")
	Example(ArrowSeries())
}
