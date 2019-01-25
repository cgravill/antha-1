package data

import (
	"reflect"
	"testing"
)

func TestFromStructs(t *testing.T) {
	type foo struct {
		Bar        int64
		Baz        []string
		unexported float64
	}
	tab := Must().NewTableFromStructs([]foo{})
	assertEqual(t, NewTable([]*Series{
		nativeSeries("Bar", []int64{}),
		nativeSeries("Baz", [][]string{}),
	}), tab, "empty")

	tab2 := Must().NewTableFromStructs([]struct{ A, B int64 }{{11, 2}, {33, 4}})
	assertEqual(t, NewTable([]*Series{
		nativeSeries("A", []int64{11, 33}),
		nativeSeries("B", []int64{2, 4}),
	}), tab2, "anonymous struct type, filled")

	s := &foo{Bar: 1}
	tab3 := Must().NewTableFromStructs([]*foo{s})
	assertEqual(t, NewTable([]*Series{
		nativeSeries("Bar", []int64{1}),
		nativeSeries("Baz", [][]string{nil}),
	}), tab3, "ptr struct type")

	_, err := NewTableFromStructs(1)
	if err == nil {
		t.Error("kind check")
	}
	_, err = NewTableFromStructs([]int{1})
	if err == nil {
		t.Error("struct check")
	}
}
func TestToStructs(t *testing.T) {
	// TODO nulls == zero type

	tab := NewTable([]*Series{
		nativeSeries("A", []int64{1, 1000}),
		nativeSeries("B", []string{"abcdef", "abcd"}),
		nativeSeries("unexported", []string{"xx", "xx"}),
		nativeSeries("Unmapped", []string{"xx", "xx"}),
		// TODO error here: nativeSeries( "C", []float64{0, 0}),
	})
	type destT struct {
		B          string
		A          int64
		Unbound    int
		unexported string
	}

	dest := []destT{}
	err := tab.ToStructs(&dest)
	if err != nil {
		t.Fatal(err)
	}
	// notice zeros set on the unexported field
	if !reflect.DeepEqual(dest, []destT{
		{A: 1, B: "abcdef"},
		{A: 1000, B: "abcd"},
	}) {
		t.Errorf("actual: %+v", dest)
	}
	roundtrip := Must().NewTableFromStructs(dest)
	// notice column order is set by struct field order
	expected := tab.Must().Project("B", "A").Extend("Unbound").Constant(0)
	assertEqual(t, expected, roundtrip, "roundtrip")

	err = tab.Must().Project("B").Rename("B", "A").ToStructs(&dest)
	if err == nil {
		t.Error("assignability check")
	}

	err = tab.ToStructs(dest)
	if err == nil {
		t.Error("ptr check")
	}

	err = tab.ToStructs(1)
	if err == nil {
		t.Error("kind check")
	}

	destPtrs := []*destT{}
	err = tab.ToStructs(&destPtrs)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(destPtrs, []*destT{
		&destT{A: 1, B: "abcdef"},
		&destT{A: 1000, B: "abcd"},
	}) {
		t.Errorf("actual ptrs: %+v", destPtrs)
	}

}
