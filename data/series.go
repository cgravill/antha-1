package data

import (
	"reflect"

	"github.com/pkg/errors"
)

type advanceable interface {
	Next() bool // false = end iteration
}

// the generic value iterator.  Pays the cost of interface pointer on each value
type iterator interface {
	advanceable
	Value() interface{} // always must be implemented
}

// Series is a named sequence of values. for larger datasets this sequence may
// be loaded lazily (eg memory map) or may even be unbounded
type Series struct {
	col ColumnName
	// typically a scalar type
	typ  reflect.Type
	read func(*seriesIterCache) iterator
	meta SeriesMeta
}

// SeriesMeta captures differing series backend capabilities
type SeriesMeta interface {
	// IsBounded = true if the Series is bounded
	IsBounded() bool
	// IsMaterialized = true if the Series is bounded and not lazy
	IsMaterialized() bool
}

// Bounded is implemented by bounded series metadata
type Bounded interface {
	SeriesMeta
	// ExactSize can return -1 if size is not known
	ExactSize() int
	// MaxSize should always return >=0
	MaxSize() int
}

// TODO ... for efficiently indexable backend
type Sliceable interface {
	Slice(start, end Index) *Series
}

// MaterializedSeriesBuilder is an interface for building materialized Series from external data source
type MaterializedSeriesBuilder interface {
	// Reserve reserves extra buffer space
	Reserve(capacity int)
	// Size returns the number of appended values
	Size() int
	// Append appends a single value
	Append(value interface{})
	// AppendNull appends nil value
	AppendNull()
	// Build constructs Series
	Build() *Series
}

func (s *Series) assignableTo(typ reflect.Type) error {
	if !s.typ.AssignableTo(typ) {
		return errors.Errorf("column %s of type %v cannot be iterated as %v", s.col, s.typ, typ)
	}
	return nil
}

func (s *Series) convertibleTo(typ reflect.Type) error {
	if !s.typ.ConvertibleTo(typ) {
		return errors.Errorf("column %s of type %v cannot be converted to %v", s.col, s.typ, typ)
	}
	return nil
}

// Cache converts a (possibly) lazy series to one that is fully materialized (currently Arrow)
func (s *Series) Cache() (*Series, error) {
	// optimizing the case when a series is alreary materialized
	if s.meta.IsMaterialized() {
		return s, nil
	}
	if !s.meta.IsBounded() {
		//TODO error
	}
	// materializing series
	return NewArrowSeriesFromSeries(s)
}
