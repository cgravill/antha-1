package parquet

import (
	"reflect"

	"github.com/antha-lang/antha/data"
	"github.com/pkg/errors"
	"github.com/xitongsys/parquet-go/ParquetFile"
	"github.com/xitongsys/parquet-go/ParquetReader"
	"github.com/xitongsys/parquet-go/parquet"
)

// ReadTable reads a data.Table from a Parquet file
func ReadTable(filePath string) (*data.Table, error) {
	// reading Parquet file metadata
	metadata, err := readMetadata(filePath)
	if err != nil {
		return nil, err
	}

	// transforming Parquet file metadata into data.Schema
	schema, err := schemaFromParquetMetadata(metadata)
	if err != nil {
		return nil, err
	}

	// creating a struct type to accomodate single Table row data
	// (a dynamic data type with parquet tags, which is required by parquet-go)
	rowType, err := rowStructFromSchema(schema)
	if err != nil {
		return nil, err
	}

	// starting building Series
	seriesBuilders, err := makeSeriesBuilders(schema)
	if err != nil {
		return nil, err
	}

	// reading from Parquet
	err = readFromParquet(filePath, rowType, func(row interface{}) {
		// appending row fields to series builders
		rowValue := reflect.ValueOf(row)
		for i, builder := range seriesBuilders {
			field := rowValue.Field(i)
			if !field.IsNil() {
				builder.Append(field.Elem().Interface())
			} else {
				builder.AppendNull()
			}
		}
	})
	if err != nil {
		return nil, err
	}

	// building all Series and composing a Table
	return makeTable(seriesBuilders), nil
}

// reads Parquet file metadata
func readMetadata(filePath string) (*parquet.FileMetaData, error) {
	// parquet file
	file, err := ParquetFile.NewLocalFileReader(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "read Parquet file '%s'", filePath)
	}
	defer file.Close() //nolint

	// reading parquet file footer
	parquetReader, err := ParquetReader.NewParquetReader(file, nil, 1)
	if err != nil {
		return nil, errors.Wrap(err, "create Parquet reader")
	}

	return parquetReader.Footer, nil
}

// series builders for storing data read from Parquet
func makeSeriesBuilders(schema *data.Schema) ([]data.MaterializedSeriesBuilder, error) {
	seriesBuilders := make([]data.MaterializedSeriesBuilder, len(schema.Columns))
	for i, column := range schema.Columns {
		seriesBuilder, err := data.NewArrowSeriesBuilder(column.Name, column.Type)
		if err != nil {
			return nil, err
		}
		seriesBuilders[i] = seriesBuilder
	}
	return seriesBuilders, nil
}

// Reads rows from Parquet file
func readFromParquet(filePath string, rowType reflect.Type, onRow func(interface{})) error {
	// for now, reading Parquet file in 1 thread, 100 rows at once
	// TODO: which parameters to use for really large datasets?
	np := 1
	batchSize := 100

	// parquet file
	file, err := ParquetFile.NewLocalFileReader(filePath)
	if err != nil {
		return errors.Wrapf(err, "open Parquet file '%s'", filePath)
	}
	defer file.Close() //nolint

	// parquet reader
	parquetReader, err := ParquetReader.NewParquetReader(file, reflect.New(rowType).Interface(), int64(np))
	if err != nil {
		return errors.Wrapf(err, "create Parquet reader '%s'", filePath)
	}
	defer parquetReader.ReadStop()

	// total number of rows
	numRows := int(parquetReader.GetNumRows())

	// type []rowType
	sliceType := reflect.SliceOf(rowType)
	// var *[]rowType
	slicePtr := reflect.New(sliceType)

	for numRows > 0 {
		rowCount := min(batchSize, numRows)
		numRows -= rowCount

		// make([]rowType, rowCount, rowCount)
		slicePtr.Elem().Set(reflect.MakeSlice(sliceType, rowCount, rowCount))

		// reading
		if err := parquetReader.Read(slicePtr.Interface()); err != nil {
			return errors.Wrap(err, "reading data from Parquet")
		}

		// callback
		slice := slicePtr.Elem()
		for i := 0; i < rowCount; i++ {
			onRow(slice.Index(i).Interface())
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// builds all Series and makes a Table containing them
func makeTable(seriesBuilders []data.MaterializedSeriesBuilder) *data.Table {
	series := make([]*data.Series, len(seriesBuilders))
	for i := range series {
		series[i] = seriesBuilders[i].Build()
	}
	return data.NewTable(series)
}
