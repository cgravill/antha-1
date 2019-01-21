package data

import ()
import "github.com/pkg/errors"
import "reflect"

/*
 * data binding between tables and go struct types
 */

// ToStructs copies to the exported struct fields by name,
// ignoring unmapped columns.
// structsPtr must be of a *[]struct{} or *[]*struct{} type
// Returns error if the struct fields are not assignable from the input data.
// TODO optional error on unmapped column
// TODO error on ambiguous column name
// TODO error on duplicate struct fields from different go namespaces
// TODO field to capture row index
func (t *Table) ToStructs(structsPtr interface{}) error {
	v := reflect.ValueOf(structsPtr)
	if v.Kind() != reflect.Ptr {
		return errors.Errorf("expecting ptr to slice, got %+v", v.Type())
	}
	vSlice := v.Elem()
	destType := vSlice.Type().Elem()
	level := 0
	for destType.Kind() == reflect.Ptr {
		destType = destType.Elem()
		level++
	}
	if destType.Kind() != reflect.Struct {
		return errors.Errorf("expecting *[]struct or *[]*struct, got %+v", v.Type())
	}

	// map column index to field
	colMap := map[int][]int{}
	for c, col := range t.Schema().Columns {
		if field, ok := destType.FieldByName(string(col.Name)); ok && field.PkgPath == "" {
			// type check
			if !col.Type.AssignableTo(field.Type) {
				return errors.Errorf("can't map column %v to struct field %v of type %+v", col, field.Name, field.Type)
			}
			colMap[c] = field.Index
		}
	}
	iter, done := t.Iter()
	defer done()
	for {
		r, more := <-iter
		if !more {
			break
		}
		destValue := reflect.New(destType).Elem()
		for i, fieldIndex := range colMap {
			destValue.FieldByIndex(fieldIndex).Set(
				reflect.ValueOf(r.Values[i].Interface()),
			)
		}
		for i := 0; i < level; i++ {
			destValue = destValue.Addr()
		}
		vSlice = reflect.Append(vSlice, destValue)
	}

	v.Elem().Set(vSlice)
	return nil
}

// NewTableFromStructs copies exported struct fields to a new table.
// structs must be of a []struct{} or []*struct{} type
// TODO customize databinding with field tags
// TODO embedded fields, anonymous fields, unexported fields?
// TODO this may be grossly inefficient... unsafe.Offsetof would be one approach to optimize
func NewTableFromStructs(structs interface{}) (*Table, error) {
	sliceSource := reflect.ValueOf(structs)
	if sliceSource.Kind() != reflect.Slice {
		return nil, errors.Errorf("expecting a slice, got a %+v", sliceSource.Type())
	}
	sourceType := sliceSource.Type().Elem()
	level := 0
	for sourceType.Kind() == reflect.Ptr {
		sourceType = sourceType.Elem()
		level++
	}
	if sourceType.Kind() != reflect.Struct {
		return nil, errors.Errorf("expecting slice of struct or *struct, got %+v", sliceSource.Type().Elem())
	}
	length := sliceSource.Len()
	// TODO can we reflectively construct slices of concrete scalar type with reflect.SliceOf?
	slices := make([][]interface{}, sourceType.NumField())
	for i := 0; i < sourceType.NumField(); i++ {
		field := sourceType.Field(i)
		if field.PkgPath == "" { //it's exported.
			slices[i] = make([]interface{}, length)
		}
	}
	for j := 0; j < length; j++ {
		rowVal := sliceSource.Index(j)
		for x := 0; x < level; x++ {
			rowVal = rowVal.Elem()
		}
		for i := 0; i < sourceType.NumField(); i++ {
			if slices[i] != nil {
				val := rowVal.Field(i)
				slices[i][j] = val.Interface()
			}
		}
	}
	// create a Series for each exported field
	series := []*Series{}
	for i := 0; i < sourceType.NumField(); i++ {
		if slices[i] != nil {

			field := sourceType.Field(i)
			ser, err := NewSliceSeries(ColumnName(field.Name), slices[i])
			if err != nil {
				return nil, err
			}
			ser.typ = field.Type
			series = append(series, ser)
		}
	}
	return NewTable(series), nil
}

// TODO needed? work on chan Row?
func (r Row) ToStruct(structPtr interface{}) error {
	return nil
}
