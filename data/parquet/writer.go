package parquet

import (
	"reflect"

	"github.com/antha-lang/antha/data"
	"github.com/pkg/errors"
	"github.com/xitongsys/parquet-go/ParquetFile"
	"github.com/xitongsys/parquet-go/ParquetWriter"
	"github.com/xitongsys/parquet-go/parquet"
)

// WriteTable writes a data.Table to a Parquet file
func WriteTable(table *data.Table, filePath string) error {
	// creating a struct type to accomodate single Table row data (a dynamic data type with parquet tags)
	schema := table.Schema()
	rowType, err := rowStructFromSchema(&schema)
	if err != nil {
		return err
	}

	// starting iterating through the table
	iter, done := table.Iter()
	defer done()

	// writing to Parquet
	return writeToParquet(filePath, rowType, func() (interface{}, error) {
		row, ok := <-iter
		if !ok {
			return nil, nil
		}
		return makeRowValue(row, rowType), nil
	})
}

// Writes rows to Parquet file
func writeToParquet(filePath string, rowType reflect.Type, rowIter func() (interface{}, error)) error {
	// Opening the file
	file, err := ParquetFile.NewLocalFileWriter(filePath)
	if err != nil {
		return errors.Wrapf(err, "opening Parquet file '%s' for writing", filePath)
	}
	defer file.Close() //nolint

	// Parquet writer and its settings
	writer, err := ParquetWriter.NewParquetWriter(file, reflect.New(rowType).Interface(), 1)
	if err != nil {
		return errors.Wrap(err, "creating Parquet writer")
	}
	writer.RowGroupSize = 128 * 1024 * 1024 //128M
	writer.CompressionType = parquet.CompressionCodec_SNAPPY

	// Writing to Parquet
	for {
		row, err := rowIter()
		if err != nil {
			return err
		}
		if row == nil {
			break
		}
		if err = writer.Write(row); err != nil {
			return err
		}
	}

	// Flush
	return writer.WriteStop()
}

// copies data.Row content into a dynamic data struct suitable for Parquet writer
func makeRowValue(row data.Row, rowType reflect.Type) interface{} {
	rowValue := reflect.New(rowType)
	// filling fields
	for i, obs := range row.Values {
		field := rowValue.Elem().Field(i)
		if obs.Interface() != nil {
			field.Set(reflect.New(field.Type().Elem()))
			field.Elem().Set(reflect.ValueOf(obs.Interface()))
		}
	}
	return rowValue.Interface()
}
