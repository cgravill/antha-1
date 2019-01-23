package data

import (
	"reflect"

	"github.com/pkg/errors"
)

type ColumnName string
type Index int

// Schema is intended as an immutable representation of table metadata
type Schema struct {
	Columns []Column
	Key     *Key
	byName  map[ColumnName][]int
}

// Key defines sorting columns (and directions)
type Key []ColumnKey

type ColumnKey struct {
	Column ColumnName
	Asc    bool
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

// IsPrefix checks if other key is a prefix of k
func (key Key) HasPrefix(other Key) bool {
	if len(other) > len(key) {
		return false
	}
	for i := range other {
		if !other[i].Equal(key[i]) {
			return false
		}
	}
	return true
}

// Equal checks if two keys are equal
func (key Key) Equal(other Key) bool {
	return len(other) == len(key) && key.HasPrefix(other)
}

// Equal checks if two key entries are equal
func (ck ColumnKey) Equal(other ColumnKey) bool {
	return ck.Column == other.Column && ck.Asc == other.Asc
}
