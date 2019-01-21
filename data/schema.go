package data

import (
	"reflect"

	"github.com/pkg/errors"
)

type ColumnName string
type Index int
type Key []ColumnName

// Schema is intended as an immutable representation of table metadata
type Schema struct {
	Columns []Column
	byName  map[ColumnName][]int
}

// Column is Series metadata
// TODO: model nullability
type Column struct {
	Name ColumnName
	Type reflect.Type
}

// Size is number of columns
func (s Schema) Size() int {
	return len(s.Columns)
}

// Equal returns true if order, name and types match
func (s Schema) Equal(other Schema) bool {
	if s.Size() != other.Size() {
		return false
	}
	for i, c := range s.Columns {
		co := other.Columns[i]
		if c.Name != co.Name {
			return false
		}
		// NB types are cached but exact equality is hard to test for, see reflect.go
		if c.Type != co.Type {
			if c.Type.Kind() != co.Type.Kind() {
				return false
			}

			if !c.Type.AssignableTo(co.Type) || !co.Type.AssignableTo(c.Type) {
				return false
			}
		}
	}
	return true
}

// Col gets the column by name, first matched
func (s Schema) Col(col ColumnName) (Column, error) {
	if index, err := s.ColIndex(col); err == nil {
		return s.Columns[index], nil
	} else {
		return Column{}, err
	}
}

// Col gets the column index by name, first matched
func (s Schema) ColIndex(col ColumnName) (int, error) {
	if cs, found := s.byName[col]; found {
		return cs[0], nil
	} else {
		return -1, errors.Errorf("no such column: %v", col)
	}
}

// TODO String()

func newSchema(series []*Series) Schema {
	schema := Schema{byName: map[ColumnName][]int{}}
	for c, s := range series {
		schema.Columns = append(schema.Columns, Column{Type: s.typ, Name: s.col})
		schema.byName[s.col] = append(schema.byName[s.col], c)
	}
	return schema
}
