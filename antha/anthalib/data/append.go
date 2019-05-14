package data

import (
	"github.com/pkg/errors"
)

type AppendSelection struct {
	tables []*Table
}

// Inner append: appends tables vertically, matches their columns by names, types and ordinals.
// The resulting table contains intersecting columns only.
// Returns an error only when called with an empty tables list.
func (s *AppendSelection) Inner() (*Table, error) {
	return s.append(appendInner)
}

// Outer append: appends tables vertically, matches their columns by names, types and ordinals.
// The resulting table contains union of columns.
// Returns an error only when called with an empty tables list.
func (s *AppendSelection) Outer() (*Table, error) {
	return s.append(appendOuter)
}

// Exact append: appends tables vertically, matches their columns by names, types and ordinals.
// All the source tables must have the same schemas (except the columns order).
func (s *AppendSelection) Exact() (*Table, error) {
	return s.append(appendExact)
}

// Positional append: appends tables vertically, matches their columns by positions.
// All the source tables must have the same schemas (except the columns names).
func (s *AppendSelection) Positional() (*Table, error) {
	return s.append(appendPositional)
}

type appendMode int

const (
	appendInner appendMode = iota
	appendOuter
	appendExact
	appendPositional
)

// append concatenates multiple tables vertically.
// Empty input tables list is considered as an error.
func (s *AppendSelection) append(mode appendMode) (*Table, error) {
	// creating the resulting table schema + projections of each source table to the resulting table
	projections, err := s.combineSchemas(mode)
	if err != nil {
		return nil, err
	}

	// empty resulting schema is an edge case: returning an empty table regardless the input tables sizes
	schema := projections[0].newSchema
	if len(schema.Columns) == 0 {
		return NewTable(), nil
	}

	// retrieving Append source tables; includes optimization for nested Appends
	sources := s.makeSources(projections)

	// append series sizes
	exactSize, maxSize := s.calcSizes()

	// series group
	group := newSeriesGroup(func() seriesGroupStateImpl {
		return &appendState{
			sources:     sources,
			sourceIndex: -1,
		}
	})

	// creating the output table series
	newSeries := make([]*Series, len(schema.Columns))
	for i, col := range schema.Columns {
		newSeries[i] = &Series{
			col:  col.Name,
			typ:  col.Type,
			read: group.read(i),
			meta: &appendSeriesMeta{
				exactSize: exactSize,
				maxSize:   maxSize,
				group:     group,
				sources:   sources,
			},
		}
	}

	return NewTable(newSeries...), nil
}

// Combines source tables schemas for Append. Returns projections of individual source tables to the output one.
func (s *AppendSelection) combineSchemas(mode appendMode) ([]nullableProjection, error) {
	// at least one input table must be supplied (regardless the mode)
	if len(s.tables) == 0 {
		return nil, errors.New("append: an empty list of tables is not supported.")
	}

	switch mode {
	case appendInner, appendOuter, appendExact:
		// name-based Append
		return s.combineSchemasByNamesAndTypes(mode)
	case appendPositional:
		// position-based Append
		return s.combineSchemasPositional()
	default:
		return nil, errors.Errorf("unknown append mode %v", mode)
	}
}

// Combines schemas for column name based Append.
func (s *AppendSelection) combineSchemasByNamesAndTypes(mode appendMode) ([]nullableProjection, error) {
	columnKeys, err := func() ([]appendColumnKey, error) {
		switch mode {
		case appendInner:
			return s.combineSchemasInner()
		case appendOuter:
			return s.combineSchemasOuter()
		case appendExact:
			return s.combineSchemasExact()
		default:
			panic(errors.Errorf("SHOULD NOT HAPPEN: unknown append mode %v", mode))
		}
	}()
	if err != nil {
		return nil, err
	}
	return s.projectByNameAndType(columnKeys), nil
}

// Combines schemas for column names based Append: returns intersecting columns only.
func (s *AppendSelection) combineSchemasInner() ([]appendColumnKey, error) {
	// initializing the columns set with the first table columns
	columnKeys := s.columnKeysFromSchema(s.tables[0].schema)

	// intersecting the columns set with the other tables ones
	for i := 1; i < len(s.tables); i++ {
		// indexing the i-th table columns
		tableColumnKeysSet := columnKeysToSet(s.columnKeysFromSchema(s.tables[i].schema))
		// intersecting joint columns set with i-th table columns (preserving the order!)
		newColumnKeys := make([]appendColumnKey, 0)
		for _, columnKey := range columnKeys {
			if _, found := tableColumnKeysSet[columnKey]; found {
				newColumnKeys = append(newColumnKeys, columnKey)
			}
		}
		columnKeys = newColumnKeys
	}

	return columnKeys, nil
}

func columnKeysToSet(columnKeys []appendColumnKey) map[appendColumnKey]bool {
	columnKeysSet := make(map[appendColumnKey]bool)
	for _, key := range columnKeys {
		columnKeysSet[key] = true
	}
	return columnKeysSet
}

// Combines schemas for column name based Append: returns all the columns from all the tables.
func (s *AppendSelection) combineSchemasOuter() ([]appendColumnKey, error) {
	columnKeys := make([]appendColumnKey, 0)
	columnKeysSet := make(map[appendColumnKey]bool)

	for _, t := range s.tables {
		for _, columnKey := range s.columnKeysFromSchema(t.schema) {
			if _, found := columnKeysSet[columnKey]; !found {
				columnKeys = append(columnKeys, columnKey)
				columnKeysSet[columnKey] = true
			}
		}
	}

	return columnKeys, nil
}

// Combines schemas for column name based Append: requires the same columns in all the tables.
func (s *AppendSelection) combineSchemasExact() ([]appendColumnKey, error) {
	// tables schemas must be identical (except columns order)
	for i := 1; i < len(s.tables); i++ {
		if err := s.compareSchemasByNamesAndTypes(s.tables[0].schema, s.tables[i].schema); err != nil {
			return nil, err
		}
	}

	// since all the schemas are identical, return the first one
	return s.columnKeysFromSchema(s.tables[0].schema), nil
}

// Compares schemas of two tables for exact name-based Append. Returns an error if they are different.
// Schemas are compared by names and types only, regardless the columns order.
func (s *AppendSelection) compareSchemasByNamesAndTypes(schema1 Schema, schema2 Schema) error {
	numColumns1, numColumns2 := schema1.NumColumns(), schema2.NumColumns()
	if numColumns1 != numColumns2 {
		return errors.Errorf("append: tables having different numbers of columns encountered (%d and %d)", numColumns1, numColumns2)
	}

	columnKeys1 := s.columnKeysFromSchema(schema1)
	columnKeysSet2 := columnKeysToSet(s.columnKeysFromSchema(schema2))
	if len(columnKeysSet2) < numColumns2 {
		panic("SHOULD NOT HAPPEN: duplicate column keys")
	}

	for _, columnKey := range columnKeys1 {
		if _, found := columnKeysSet2[columnKey]; !found {
			return errors.Errorf("append: the column (name='%s', type='%v', ordinal='%d') is missing in the other table", columnKey.col.Name, columnKey.col.Type, columnKey.ordinal)
		}
	}
	return nil
}

// Makes source table projections given the output table schema.
func (s *AppendSelection) projectByNameAndType(columnKeys []appendColumnKey) []nullableProjection {
	// transforming column keys back to a Schema
	columns := make([]Column, len(columnKeys))
	for i := range columnKeys {
		columns[i] = columnKeys[i].col
	}
	outputSchema := *NewSchema(columns)

	// creating an index "columnKey -> position of this column"
	columnsKeyToPos := make(map[appendColumnKey]int)
	for i, key := range columnKeys {
		columnsKeyToPos[key] = i
	}

	// making the projections
	projections := make([]nullableProjection, len(s.tables))
	for i, t := range s.tables {
		projection := newNullableProjection(outputSchema)
		tableColumnKeys := s.columnKeysFromSchema(t.schema)
		for j, key := range tableColumnKeys {
			if pos, found := columnsKeyToPos[key]; found {
				projection.newToOld[pos] = j
			}
		}
		projections[i] = projection
	}

	return projections
}

// nullableProjection is an extended version of `projection` which supports inserting null columns.
type nullableProjection struct {
	newToOld  []int  // new column index -> old column index; unlike projection.newToOld, can contain `-1`s
	newSchema Schema // schema of the new table; newToOld[i] == -1 means that the i-th column will be a null constant column of the name and type defined by this schema
}

// Creates an empty nullable projection from a given schema.
func newNullableProjection(schema Schema) nullableProjection {
	newToOld := make([]int, schema.NumColumns())
	for i := range newToOld {
		newToOld[i] = -1
	}
	return nullableProjection{
		newToOld:  newToOld,
		newSchema: schema,
	}
}

// Creates an identity nullable projection from a given schema.
func newIdentityNullableProjection(schema Schema) nullableProjection {
	newToOld := make([]int, schema.NumColumns())
	for i := range newToOld {
		newToOld[i] = i
	}
	return nullableProjection{
		newToOld:  newToOld,
		newSchema: schema,
	}
}

// Projects a Table using a nullableProjection.
// NB: actually, performs Project+ExtendByConstant, so may need rewriting in case some day we rethink the way Project and/or Extend works.
func (p nullableProjection) project(t *Table) *Table {
	series := make([]*Series, len(p.newToOld))
	for pNew, pOld := range p.newToOld {
		if pOld == -1 {
			col := p.newSchema.Columns[pNew]
			series[pNew] = NewNullSeries(col.Name, col.Type)
		} else {
			series[pNew] = t.series[pOld]
		}
	}
	return NewTable(series...)
}

// The key which is used for columns matching in Append modes which match columns by name and type.
type appendColumnKey struct {
	col     Column // column name and type
	ordinal int    // column ordinal among the columns with the same name and type
}

// Returns the list of the table columns in the form of {(name, type, ordinal among the columns with the same name and type)}.
func (s *AppendSelection) columnKeysFromSchema(schema Schema) []appendColumnKey {
	keys := make([]appendColumnKey, len(schema.Columns))
	// column name, columns type -> number of columns with such name and type encountered by now
	columnToOrdinal := make(map[Column]int)
	for i, col := range schema.Columns {
		ordinal := columnToOrdinal[col]
		keys[i] = appendColumnKey{
			col:     col,
			ordinal: ordinal,
		}
		columnToOrdinal[col] = ordinal + 1
	}
	return keys
}

// Combines schemas for positional Append - i.e. just ensures they are identical.
func (s *AppendSelection) combineSchemasPositional() ([]nullableProjection, error) {
	// tables schemas must be identical (except column names)
	for i := 1; i < len(s.tables); i++ {
		if err := s.compareSchemasPositional(s.tables[0].schema, s.tables[i].schema); err != nil {
			return nil, err
		}
	}

	// in case of positional append, all the tables projections are identity projections
	projections := make([]nullableProjection, len(s.tables))
	for i := range projections {
		projections[i] = newIdentityNullableProjection(s.tables[0].schema)
	}

	return projections, nil
}

// Compares schemas of two tables for positional Append, returns an error if they are different.
// Schemas are compared by types only, regardless columns names.
func (s *AppendSelection) compareSchemasPositional(schema1 Schema, schema2 Schema) error {
	numColumns1, numColumns2 := schema1.NumColumns(), schema2.NumColumns()
	if numColumns1 != numColumns2 {
		return errors.Errorf("append: tables having different numbers of columns encountered (%d and %d)", numColumns1, numColumns2)
	}
	for i := range schema1.Columns {
		type1, type2 := schema1.Columns[i].Type, schema2.Columns[i].Type
		if type1 != type2 {
			return errors.Errorf("append: tables columns #%d having different types (%v and %v)", i, type1, type2)
		}
	}
	return nil
}

// Creates a list of Append source tables.
func (s *AppendSelection) makeSources(projections []nullableProjection) []*Table {
	newSources := make([]*Table, 0, len(s.tables))
	for i, t := range s.tables {
		// optimization: if an Append source table is itself a product of another `Append` call, then using its sources directly
		// (in order to simplify iteration over the resulting table)
		tSources, extracted := s.extractSources(t)
		if !extracted {
			// otherwise, using the source table itself
			tSources = []*Table{t}
		}

		// applying the corresponding projection to the sources
		for _, source := range tSources {
			newSource := projections[i].project(source)
			newSources = append(newSources, newSource)
		}
	}
	return newSources
}

// If the table is itself a result of Append, then extracts its source tables. Otherwise, returns false.
func (s *AppendSelection) extractSources(table *Table) ([]*Table, bool) {
	for _, series := range table.series {
		// if some series is not Append series, the table is not the result of a single Append
		meta, ok := series.meta.(*appendSeriesMeta)
		if !ok {
			return nil, false
		}
		// if some of Append series do not share the same appendTableMeta, the table is not the result of a single Append
		if meta.group != table.series[0].meta.(*appendSeriesMeta).group {
			return nil, false
		}
	}
	return table.series[0].meta.(*appendSeriesMeta).sources, true
}

func (s *AppendSelection) calcSizes() (exactSize int, maxSize int) {
	for _, t := range s.tables {
		exactSizeT, maxSizeT, err := seriesSize(t.series)
		if err != nil {
			panic(errors.Wrap(err, "SHOULD NOT HAPPEN"))
		}
		exactSize = addSizes(exactSize, exactSizeT)
		maxSize = addSizes(maxSize, maxSizeT)
	}
	return
}

// adds two sizes; if one of them is undefined, the sum is undefined too
func addSizes(size1 int, size2 int) int {
	const undefinedSize = -1
	if size1 == undefinedSize || size2 == undefinedSize {
		return undefinedSize
	}
	return size1 + size2
}

// appendTables resulting series metadata
type appendSeriesMeta struct {
	exactSize int
	maxSize   int
	// needed for optimization purposes only (extractSources)
	group   *seriesGroup
	sources []*Table
}

func (m *appendSeriesMeta) IsMaterialized() bool { return false }
func (m *appendSeriesMeta) ExactSize() int       { return m.exactSize }
func (m *appendSeriesMeta) MaxSize() int         { return m.maxSize }

// Common state of several Append series. They share the source table iterator (including series iterators cache).
type appendState struct {
	sources     []*Table       // source tables
	sourceIndex int            // current source table
	sourceIter  *tableIterator // current source table iterator
}

func (st *appendState) Next() bool {
	for {
		// first trying to advance the current table iterator
		if st.sourceIndex != -1 && st.sourceIter.Next() {
			return true
		}
		// moving to the next table if possible
		if st.sourceIndex+1 == len(st.sources) {
			return false
		}
		st.sourceIndex++
		st.sourceIter = newTableIterator(st.sources[st.sourceIndex].series)
	}
}

func (st *appendState) Value(colIndex int) interface{} {
	return st.sourceIter.colReader[colIndex].Value()
}
