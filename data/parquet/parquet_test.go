package parquet

import (
	"io/ioutil"
	"math"
	"os"
	"testing"

	"github.com/antha-lang/antha/data"
)

func TestParquet(t *testing.T) {
	// create a Table from Arrow Series
	column1 := data.Must().NewArrowSeriesFromSlice("bool_column", []bool{true, true, false, false, true}, nil)
	column2 := data.Must().NewArrowSeriesFromSlice("int64_column", []int64{10, 10, 30, -1, 5}, []bool{true, true, true, false, true})
	column3 := data.Must().NewArrowSeriesFromSlice("float32_column", []float64{1.5, 2.5, 3.5, math.NaN(), 5.5}, []bool{true, true, true, false, true})
	column4 := data.Must().NewArrowSeriesFromSlice("string_column", []string{"", "aa", "xx", "aa", ""}, nil)
	table := data.NewTable([]*data.Series{column1, column2, column3, column4})

	// write Table to Parquet
	fileName := parquetFileName(t)
	defer os.Remove(fileName)

	if err := WriteTable(table, fileName); err != nil {
		t.Errorf("write table: %s", err)
	}

	// read Table to Parquet
	readTable, err := ReadTable(fileName)
	if err != nil {
		t.Errorf("read table: %s", err)
	}

	assertEqual(t, table, readTable, "tables are different after serialization")
}

func parquetFileName(t *testing.T) string {
	f, err := ioutil.TempFile("", "table*.parquet")
	if err != nil {
		t.Errorf("create temp file: %s", err)
	}
	defer f.Close() //nolint
	return f.Name()
}

func assertEqual(t *testing.T, expected, actual *data.Table, msg string) {
	if !actual.Equal(expected) {
		t.Error(msg)
		t.Log("actual", actual.Head(20).ToRows())
	}
}
