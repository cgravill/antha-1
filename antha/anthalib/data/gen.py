#!/usr/bin/env python

# types with arrow and native adapter support
TYPE_SUPPORT = [
	{'Raw':'float64', 'Type':'Float64', 'Zero':'float64(0)', 'ArrowArrayType':'Float64', 'ArrowRawType':'float64'},
	{'Raw':'int64', 'Type':'Int64', 'Zero':'int64(0)', 'ArrowArrayType':'Int64', 'ArrowRawType':'int64'},
	{'Raw':'string', 'Type':'String', 'Zero':'""', 'ArrowArrayType':'String', 'ArrowRawType':'string'},
	{'Raw':'bool', 'Type':'Bool', 'Zero':'false', 'ArrowArrayType':'Boolean', 'ArrowRawType':'bool'},
	# TODO: make decision on how many time/timestamp types we need
	{'Raw':'TimestampMillis', 'Type':'TimestampMillis', 'Zero':'TimestampMillis(0)', 'ArrowArrayType':'Timestamp', 'ArrowRawType':'arrow.Timestamp', 'TimeUnit':"Millisecond"},
	{'Raw':'TimestampMicros', 'Type':'TimestampMicros', 'Zero':'TimestampMicros(0)', 'ArrowArrayType':'Timestamp', 'ArrowRawType':'arrow.Timestamp', 'TimeUnit':"Microsecond"},
]


# types having custom extended comparator
CUSTOM_COMPARE = ['string', 'bool']

TPL = {}

TPL['series.gen.go'] = r'''
package data
// Code generated by gen.py. DO NOT EDIT.

/*
 * 'iter<Type'> are iterator specializations for potentially no-copy, boxed values.
 *
 * The 'as<Type>' types are fallbacks for when the underlying series is dynamic.
 */

{% for t in env['TYPE_SUPPORT'] -%}

// box{{ t['Type'] }} represents a nullable {{ t['Raw'] }} value
type box{{ t['Type'] }} interface {
	{{ t['Type'] }}() ({{ t['Raw'] }}, bool) // returns false = nil
}

// iter{{ t['Type'] }} iterates over nullable {{ t['Raw'] }} values
type iter{{ t['Type'] }} interface {
	advanceable
	box{{ t['Type'] }}
}

// iterate{{ t['Type'] }} is a fallback to convert dynamic series to static iterator type.
// an error is returned if the series' declared type is not assignable to {{ t['Raw'] }}
func (s *Series) iterate{{ t['Type'] }}(iter iterator) (iter{{ t['Type'] }}, error) {
	if cast, ok := iter.(iter{{ t['Type'] }}); ok {
		return cast, nil
	}
	if err := s.assignableTo(type{{ t['Type'] }}); err != nil {
		return nil, err
	}
	return &as{{ t['Type'] }}{iterator: iter}, nil
}

type as{{ t['Type'] }} struct {
	iterator
}

func (a *as{{ t['Type'] }}) {{ t['Type'] }}() ({{ t['Raw'] }}, bool) {
	v := a.iterator.Value()
	if v == nil {
		return {{ t['Zero'] }}, false
	}
	return v.({{ t['Raw'] }}), true
}
{% endfor %}
'''

TPL['extend.gen.go']=r'''
package data
// Code generated by gen.py. DO NOT EDIT.

import (
	"github.com/pkg/errors"
)

{% for t in env['TYPE_SUPPORT'] -%}
// {{ t['Type'] }} adds a {{ t['Raw'] }} col using {{ t['Raw'] }} inputs.  Null on any null inputs.
// Returns error if any column cannot be assigned to {{ t['Raw'] }}; no conversions are performed.
func (e *ExtendOn) {{ t['Type'] }}(f func(v ...{{ t['Raw'] }}) {{ t['Raw'] }}) (*Table, error) {
	typ := type{{ t['Type'] }}
	inputs, err := e.inputs(typ)
	if err != nil {
		return nil, err
	}
	return newFromSeries(append(append([]*Series(nil), e.extension.t.series...), &Series{
		col: e.extension.newCol,
		typ: typ,
		meta: e.meta,
		read: func(cache *seriesIterCache) iterator {
			colReader := make([]iter{{ t['Type'] }}, len(inputs))
			var err error
			for i, ser := range inputs {
				iter := cache.Ensure(ser)
				colReader[i], err = ser.iterate{{ t['Type'] }}(iter) // note colReader[i] is not itself in the cache!
				if err != nil {
					panic(errors.Wrapf(err, "SHOULD NOT HAPPEN; when extending new column %q", e.extension.newCol))
				}
			}
			// end when table exhausted
			e.extension.extensionSource(cache)
			return &extend{{ t['Type'] }}{f: f, source: colReader}
		}},
	), e.extension.t.sortKey...), nil
}

var _ iter{{ t['Type'] }} = (*extend{{ t['Type'] }})(nil)

type extend{{ t['Type'] }} struct {
	f      func(v ...{{ t['Raw'] }}) {{ t['Raw'] }}
	source []iter{{ t['Type'] }}
}

func (x *extend{{ t['Type'] }}) Next() bool {
	return true
}

func (x *extend{{ t['Type'] }}) Value() interface{} {
	v, ok := x.{{ t['Type'] }}()
	if !ok {
		return nil
	}
	return v
}
func (x *extend{{ t['Type'] }}) {{ t['Type'] }}() ({{ t['Raw'] }}, bool) {
	args := make([]{{ t['Raw'] }}, len(x.source))
	var ok bool
	for i, s := range x.source {
		args[i], ok = s.{{ t['Type'] }}()
		if !ok {
			return {{ t['Zero'] }}, false
		}
	}
	v := x.f(args...)
	return v, true
}

// {{ t['Type'] }} adds a {{ t['Raw'] }} col using {{ t['Raw'] }} inputs.  Null on any null inputs.
// Panics on error.
func (m *MustExtendOn) {{ t['Type'] }}(f func(v ...{{ t['Raw'] }}) {{ t['Raw'] }}) *Table {
	t, err := m.ExtendOn.{{ t['Type'] }}(f)
	handle(err)
	return t
}

// Interface{{ t['Type'] }} adds a {{ t['Raw'] }} col using arbitrary (interface{}) inputs.
func (e *ExtendOn) Interface{{ t['Type'] }}(f func(v ...interface{}) ({{ t['Raw'] }}, bool)) (*Table, error) {
	projection, err := newProjection(e.extension.t.schema, e.inputCols...)
	if err != nil {
		return nil, err
	}

	return newFromSeries(append(append([]*Series(nil), e.extension.t.series...), &Series{
		col: e.extension.newCol,
		typ: type{{ t['Type'] }},
		meta: e.meta,
		read: func(cache *seriesIterCache) iterator {
			colReader := make([]iterator, len(projection.newToOld))
			for new, old := range projection.newToOld {
				colReader[new] = cache.Ensure(e.extension.t.series[old])
			}
			// end when table exhausted
			//e.extension.extensionSource(cache)
			return &extendInterface{{ t['Type'] }}{f: f, source: colReader}
		}},
	), e.extension.t.sortKey...), nil
}

var _ iter{{ t['Type'] }} = (*extendInterface{{ t['Type'] }})(nil)

type extendInterface{{ t['Type'] }} struct {
	f      func(v ...interface{}) ({{ t['Raw'] }}, bool)
	source []iterator
}

func (x *extendInterface{{ t['Type'] }}) Next() bool {
	return true
}

func (x *extendInterface{{ t['Type'] }}) Value() interface{} {
	v, ok := x.{{ t['Type'] }}()
	if !ok {
		return nil
	}
	return v
}

func (x *extendInterface{{ t['Type'] }}) {{ t['Type'] }}() ({{ t['Raw'] }}, bool) {
	args := make([]interface{}, len(x.source))
	for i, s := range x.source {
		args[i] = s.Value()
	}
	return x.f(args...)
}

// Interface{{ t['Type'] }} adds a {{ t['Raw'] }} col using arbitrary (interface{}) inputs.
// Panics on error.
func (m *MustExtendOn) Interface{{ t['Type'] }}(f func(v ...interface{}) ({{ t['Raw'] }}, bool)) *Table {
	t, err := m.ExtendOn.Interface{{ t['Type'] }}(f)
	handle(err)
	return t
}

{% endfor %}
'''

TPL['filter.gen.go']=r'''
package data
// Code generated by gen.py. DO NOT EDIT.

import "github.com/pkg/errors"

{% for t in env['TYPE_SUPPORT'] -%}

// Match{{ t['Type'] }} implements a filter on {{ t['Raw'] }} columns.
type Match{{ t['Type'] }} func(...{{ t['Raw'] }}) bool

// {{ t['Type'] }} matches the named column values as {{ t['Raw'] }} arguments. 
// If any column is nil the filter is automatically false.
// If given any SchemaAssertions, they are called now and may have side effects.
func (o *FilterOn) {{ t['Type'] }}(fn Match{{ t['Type'] }}, assertions ...SchemaAssertion) (*Table, error) {
	if err := o.checkSchema(type{{ t['Type'] }}, assertions...); err != nil {
		return nil, errors.Wrapf(err, "can't filter %+v with %+v", o.t, fn)
	}

	projection := mustNewProjection(o.t.schema, o.cols...)

	matchGen := func() rawMatch {
		return func(r raw) bool {
			matchVals := make([]{{ t['Raw'] }}, len(o.cols))
			for new, old := range projection.newToOld {
				val := r[old]
				if val == nil {
					return false
				}
				matchVals[new] = val.({{ t['Raw'] }})
			}
			return fn(matchVals...)
		}
	}

	return filterTable(matchGen, o.t), nil
}

// {{ t['Type'] }} matches the named column values as {{ t['Raw'] }} arguments.
func (o *MustFilterOn) {{ t['Type'] }}(m Match{{ t['Type'] }}, assertions ...SchemaAssertion) *Table {
	t, err := o.FilterOn.{{ t['Type'] }}(m, assertions...)
	handle(err)
	return t
}

{% endfor %}
'''

TPL['native.gen.go']=r'''
package data
// Code generated by gen.py. DO NOT EDIT.

func (m *nativeSeriesMeta) read(cache *seriesIterCache) iterator {
	switch m.rValue.Type().Elem() {
	{% for t in env['TYPE_SUPPORT'] -%}
	case type{{ t['Type'] }}:
		return m.read{{ t['Type'] }}(cache)
	{% endfor -%}
	default:
		return m.fallbackRead(cache)
	}
}

{% for t in env['TYPE_SUPPORT'] -%}

// {{ t['Raw'] }}

// typed series builder

type nativeSeriesBuilder{{ t['Type'] }} struct {
	column  Column
	data    []{{ t['Raw'] }}
	notNull []bool
}

func newNativeSeriesBuilder{{ t['Type'] }}(columnName ColumnName) *nativeSeriesBuilder{{ t['Type'] }} {
	return &nativeSeriesBuilder{{ t['Type'] }}{
		column:  Column{
			Name: columnName,
			Type: type{{ t['Type'] }},
		},
		data:    []{{ t['Raw'] }}{},
		notNull: []bool{},
	}
}

func (b *nativeSeriesBuilder{{ t['Type'] }}) Column() Column {
	return b.column
}

func (b *nativeSeriesBuilder{{ t['Type'] }}) Reserve(capacity int) {
	if capacity > cap(b.data) {
		size := len(b.data)

		newData := make([]{{ t['Raw'] }}, size, capacity)
		copy(newData, b.data)
		b.data = newData

		newNotNull := make([]bool, size, capacity)
		copy(newNotNull, b.notNull)
		b.notNull = newNotNull
	}
}

func (b *nativeSeriesBuilder{{ t['Type'] }}) Size() int {
	return len(b.data)
}

func (b *nativeSeriesBuilder{{ t['Type'] }}) Append(value interface{}) {
	if value != nil {
		b.Append{{ t['Type'] }}(value.({{ t['Raw'] }}), true)
	} else {
		b.Append{{ t['Type'] }}({{ t['Zero'] }}, false)
	}
}

func (b *nativeSeriesBuilder{{ t['Type'] }}) Append{{ t['Type'] }}(value {{ t['Raw'] }}, notNull bool) {
	b.data = append(b.data, value)
	b.notNull = append(b.notNull, notNull)
}

func (b *nativeSeriesBuilder{{ t['Type'] }}) Build() *Series {
	return mustNewNativeSeriesFromSlice(b.column.Name, b.data, b.notNull)
}

var _ seriesBuilder{{ t['Type'] }} = (*nativeSeriesBuilder{{ t['Type'] }})(nil)

// typed iterator

func (m *nativeSeriesMeta) read{{ t['Type'] }}(_ *seriesIterCache) iterator {
	return &nativeSeriesIter{{ t['Type'] }}{
		data:    m.rValue.Interface().([]{{ t['Raw'] }}),
		notNull: m.notNull,
		pos:     -1,
	}
}

type nativeSeriesIter{{ t['Type'] }} struct {
	data    []{{ t['Raw'] }}
	notNull notNullMask
	pos     int
}

func (i *nativeSeriesIter{{ t['Type'] }}) Next() bool {
	i.pos++
	return i.pos < len(i.data)
}

func (i *nativeSeriesIter{{ t['Type'] }}) {{ t['Type'] }}() ({{ t['Raw'] }}, bool) {
	return i.data[i.pos], i.notNull.Test(i.pos)
}

func (i *nativeSeriesIter{{ t['Type'] }}) Value() interface{} {
	if val, ok := i.{{ t['Type'] }}(); ok {
		return val
	} else {
		return nil
	}
}

var _ iter{{ t['Type'] }} = (*nativeSeriesIter{{ t['Type'] }})(nil)

{% endfor %}
'''

TPL['arrow.gen.go']=r'''
package data
// Code generated by gen.py. DO NOT EDIT.

import (
	"reflect"

	"github.com/apache/arrow/go/arrow"
	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/pkg/errors"
)

// Series implemented on the top of Apache Arrow.

// newArrowSeriesFromSlice converts a slice of scalars to a new (Arrow-based) Series.
// notNull denotes elements set to null; it is optional and can be set to nil.
// Only a closed list of primitive data types is supported.
func newArrowSeriesFromSlice(col ColumnName, values interface{}, notNull []bool) (*Series, error) {
	switch typedValues := values.(type) {
	{% for t in env['TYPE_SUPPORT'] -%}
	case []{{ t['Raw'] }}:
		return newArrowSeriesFromSlice{{ t['Type'] }}(col, typedValues, notNull), nil
	{% endfor -%}
	default:
		return nil, errors.Errorf("The data type %v is not supported, expecting slice of supported primitive types", reflect.TypeOf(values))
	}
}

{% for t in env['TYPE_SUPPORT'] -%}

// {{ t['Raw'] }}

type arrowSeriesBuilder{{ t['Type'] }} struct {
	builder *array.{{ t['ArrowArrayType'] }}Builder
	column Column
}

func newArrowSeriesBuilder{{ t['Type'] }}(col ColumnName) *arrowSeriesBuilder{{ t['Type'] }} {
	return &arrowSeriesBuilder{{ t['Type'] }} {
		{% if t['ArrowArrayType'] != 'Timestamp' -%}
		builder: array.New{{ t['ArrowArrayType'] }}Builder(memory.DefaultAllocator),
		{% else -%}
		builder: array.New{{ t['ArrowArrayType'] }}Builder(memory.DefaultAllocator, &arrow.TimestampType{Unit: arrow.{{ t['TimeUnit'] }}}),
		{% endif -%}
		column:  Column{
			Name: col,
			Type: type{{ t['Type'] }},
		},
	}
}

func (b *arrowSeriesBuilder{{ t['Type'] }}) Column() Column       { return b.column }
func (b *arrowSeriesBuilder{{ t['Type'] }}) Reserve(capacity int) { b.builder.Reserve(capacity) }
func (b *arrowSeriesBuilder{{ t['Type'] }}) Size() int            { return b.builder.Len() }

func (b *arrowSeriesBuilder{{ t['Type'] }}) Append(value interface{}) {
	if value == nil {
		b.builder.AppendNull()
		return
	}

	typedValue := value.({{ t['Raw'] }})
	{% if t['Raw'] == t['ArrowRawType'] -%}
	b.builder.Append(typedValue)
	{% else -%}
	b.builder.Append({{ t['ArrowRawType'] }}(typedValue))
	{%- endif -%}
}

func (b *arrowSeriesBuilder{{ t['Type'] }}) Append{{ t['Type'] }}(value {{ t['Raw'] }}, notNull bool) {
	if !notNull {
		b.builder.AppendNull()
		return
	}
	b.builder.Append({%- if t['Raw'] == t['ArrowRawType'] -%}value{%- else -%}{{ t['ArrowRawType'] }}(value){%- endif -%})
}

func (b *arrowSeriesBuilder{{ t['Type'] }}) Build() *Series {
	metadata := &arrowSeriesMeta{values: b.builder.New{{ t['ArrowArrayType'] }}Array()}
	return &Series{
		typ:  type{{ t['Type'] }},
		col:  b.column.Name,
		read: metadata.read{{ t['Type'] }},
		meta: metadata,
	}
}

var _ seriesBuilder{{ t['Type'] }} = (*arrowSeriesBuilder{{ t['Type'] }})(nil)

func newArrowSeriesFromSlice{{ t['Type'] }}(col ColumnName, values []{{ t['Raw'] }}, mask []bool) *Series {
	builder := newArrowSeriesBuilder{{ t['Type'] }}(col)
	builder.Reserve(len(values))

	for i := range values {
		builder.Append{{ t['Type'] }}(values[i], mask == nil || mask[i])
	}

	return builder.Build()
}

func (m *arrowSeriesMeta) read{{ t['Type'] }}(_ *seriesIterCache) iterator {
	return &arrowSeriesIter{{ t['Type'] }}{
		values: m.values.(*array.{{ t['ArrowArrayType'] }}),
		pos:    -1,
	}
}

type arrowSeriesIter{{ t['Type'] }} struct {
	values *array.{{ t['ArrowArrayType'] }}
	pos int
}

func (i *arrowSeriesIter{{ t['Type'] }}) Next() bool {
	i.pos++
	return i.pos < i.values.Len()
}

func (i *arrowSeriesIter{{ t['Type'] }}) {{ t['Type'] }}() ({{ t['Raw'] }}, bool) {
	if !i.values.IsNull(i.pos) {
		return {% if t['Raw'] == t['ArrowRawType'] -%}i.values.Value(i.pos){%- else -%}{{ t['Raw'] }}(i.values.Value(i.pos)){%- endif -%}, true
	} else {
		return {{ t['Zero'] }}, false
	}
}

func (i *arrowSeriesIter{{ t['Type'] }}) Value() interface{} {
	if val, ok := i.{{ t['Type'] }}(); ok {
		return val
	} else {
		return nil
	}
}

var _ iterator = (*arrowSeriesIter{{ t['Type'] }})(nil)
var _ iter{{ t['Type'] }} = (*arrowSeriesIter{{ t['Type'] }})(nil)

{% endfor %}
'''

TPL['row.gen.go']=r'''
package data
// Code generated by gen.py. DO NOT EDIT.

{% for t in env['TYPE_SUPPORT'] -%}

// {{ t['Raw'] }}

// Must{{ t['Type'] }} extracts a value of type {{ t['Raw'] }} from an observation. Panics on error.
func (o Observation) Must{{ t['Type'] }}() {{ t['Raw'] }} {
	return o.value.({{ t['Raw'] }})
}

// {{ t['Type'] }} extracts a value of type {{ t['Raw'] }} from an observation. Returns false on error.
func (o Observation) {{ t['Type'] }}() ({{ t['Raw'] }}, bool) {
	if o.IsNull() {
		return {{ t['Zero'] }}, false
	} else {
		return o.Must{{ t['Type'] }}(), true
	}
}

{% endfor %}
'''

TPL['types.gen.go']=r'''
package data

import "reflect"

// Code generated by gen.py. DO NOT EDIT.

// registry for well-known scalar types

{% for t in env['TYPE_SUPPORT'] -%}
var type{{ t['Type'] }} reflect.Type = reflect.TypeOf({{ t['Zero'] }})
{% endfor %}

var typeSupport map[reflect.Type]string = map[reflect.Type]string {
	{% for t in env['TYPE_SUPPORT']%}type{{ t['Type'] }}: "{{ t['Raw'] }}",
	{% endfor %}
}

var typeSupportByName map[string]reflect.Type = map[string]reflect.Type {
	{% for t in env['TYPE_SUPPORT']%}"{{ t['Raw'] }}": type{{ t['Type'] }},
	{% endfor %}
}
'''


TPL['sort.gen.go']=r'''
package data
// Code generated by gen.py. DO NOT EDIT.

import "github.com/pkg/errors"

// newNativeCompareFunc creates a function to compare elements of the given native Series.
func newNativeCompareFunc(nativeSeries *Series, asc bool) (compareFunc, error) {
	meta, ok := nativeSeries.meta.(*nativeSeriesMeta)
	if !ok {
		panic(errors.Errorf("series %+v is not native", nativeSeries))
	}

	// TODO: more optimal comparators for non-nullable columns
	switch nativeSeries.typ { 
	{% for t in env['TYPE_SUPPORT'] -%}
	case type{{ t['Type'] }}:
		return newNativeCompareFunc{{ t['Type'] }}(meta, asc), nil
	{% endfor -%}
	default:
		// Currently we don't have a generic native series compare function.
		// However, it is possible to write a reflective one - at least for certain Kinds.
		return nil, errors.Errorf("The data type %+v is not supported, expecting a series of some supported primitive type", nativeSeries.typ)
	}
}

// newNativeSwapFunc creates a function to swap elements of the given native Series.
func newNativeSwapFunc(nativeSeries *Series) swapFunc {
	meta, ok := nativeSeries.meta.(*nativeSeriesMeta)
	if !ok {
		panic(errors.Errorf("series %+v is not native", nativeSeries))
	}

	switch nativeSeries.typ { 
	{% for t in env['TYPE_SUPPORT'] -%}
	case type{{ t['Type'] }}:
		return newNativeSwapFunc{{ t['Type'] }}(meta)
	{% endfor -%}
	default:
		// a fallback swap func generator (very slow!)
		return newNativeSwapFuncGeneric(meta)
	}
}

{% for t in env['TYPE_SUPPORT'] -%}
// {{ t['Raw'] }}

func newNativeCompareFunc{{ t['Type'] }}(nativeMeta *nativeSeriesMeta, asc bool) compareFunc {
	data := nativeMeta.rValue.Interface().([]{{ t['Raw'] }})
	notNull := nativeMeta.notNull

	return func(i, j int) int {
		return compare{{ t['Type'] }}(data[i], notNull.Test(i), data[j], notNull.Test(j), asc)
	}
}

func compare{{ t['Type'] }}(val1 {{ t['Raw'] }}, notNull1 bool, val2 {{ t['Raw'] }}, notNull2 bool, asc bool) int {
	result, ok := compareNulls(notNull1, notNull2)
	if !ok {
		result = rawCompare{{ t['Type'] }}(val1, val2)
	}
	return applyAsc(result, asc)
}

{% if t['Raw'] not in env['CUSTOM_COMPARE'] %}
func rawCompare{{ t['Type'] }}(val1, val2 {{ t['Raw'] }}) int {
	switch {
	case val1 < val2:
		return -1
	case val1 > val2:
		return 1
	default:
		return 0
	}
}

{% endif %}

func newNativeSwapFunc{{ t['Type'] }}(nativeMeta *nativeSeriesMeta) swapFunc {
	data := nativeMeta.rValue.Interface().([]{{ t['Raw'] }})
	notNull := nativeMeta.notNull
	return func(i, j int) {
		data[i], data[j] = data[j], data[i]
		notNull.Swap(i, j)
	}
}
{% endfor %}
'''

TPL['materialize.gen.go']=r'''
package data
// Code generated by gen.py. DO NOT EDIT.

import (
	"reflect"

	"github.com/pkg/errors"
)

func newSeriesBuilder(col ColumnName, typ reflect.Type, mode materializedType) (seriesBuilder, error) {
	switch typ {
		{% for t in env['TYPE_SUPPORT'] -%}
		case type{{ t['Type'] }}:
			return newSeriesBuilder{{ t['Type'] }}(col, mode), nil
		{% endfor -%}
		default:
			return newFallbackSeriesBuilder(col, typ, mode)
	}
}

func newSeriesCopier(s *Series, iter iterator, mode materializedType) (seriesCopier, error) {
	switch s.typ {
		{% for t in env['TYPE_SUPPORT'] -%}
		case type{{ t['Type'] }}:
			return newSeriesCopier{{ t['Type'] }}(s, iter, mode), nil
		{% endfor -%}
		default:
			return newFallbackSeriesCopier(s, iter, mode)
	}
}

{% for t in env['TYPE_SUPPORT'] -%}
// {{ t['Raw'] }}

// seriesBuilder{{ t['Type'] }} is a typed series builder for {{ t['Raw'] }}
type seriesBuilder{{ t['Type'] }} interface {
	seriesBuilder
	Append{{ t['Type'] }}(value {{ t['Raw'] }}, notNull bool)
}

func newSeriesBuilder{{ t['Type'] }}(col ColumnName, mode materializedType) seriesBuilder{{ t['Type'] }} {
	switch mode {
	case nativeSeries:
		return newNativeSeriesBuilder{{ t['Type'] }}(col)
	case arrowSeries:
		return newArrowSeriesBuilder{{ t['Type'] }}(col)
	default:
		panic(errors.Errorf("unknown materialized series type %v", mode))
	}
}

// copies a series: reads values from a source column iterator and writes them to a target column builder
type seriesCopier{{ t['Type'] }} struct {
	seriesBuilder{{ t['Type'] }}
	iter iter{{ t['Type'] }}
}

func newSeriesCopier{{ t['Type'] }}(s *Series, iter iterator, mode materializedType) *seriesCopier{{ t['Type'] }} {
	// typed source series iterator
	typedIter, err := s.iterate{{ t['Type'] }}(iter)
	if err != nil {
		panic(errors.Wrapf(err, "SHOULD NOT HAPPEN: column %s is not {{ t['Type'] }}", s.col))
	}

	// typed destination series builder
	builder := newSeriesBuilder{{ t['Type'] }}(s.col, mode)

	return &seriesCopier{{ t['Type'] }}{
		seriesBuilder{{ t['Type'] }}: builder,
		iter: typedIter,
	}
}

func (c *seriesCopier{{ t['Type'] }}) CopyValue() { c.Append{{ t['Type'] }}(c.iter.{{ t['Type'] }}()) }

{% endfor %}
'''

import jinja2
import subprocess

def write_tpl(tpl, name):
	template = jinja2.Template(tpl)
	with open(name, 'wb') as f:
		f.write(template.render(env=globals()))
	subprocess.call(["gofmt", "-s", "-w", name])

if __name__ == "__main__":
	for (f, tpl) in TPL.iteritems():
		write_tpl(tpl, f)