package data

type readRow struct {
	cols          []*ColumnName
	colReader     []iterator
	index         int
	iteratorCache seriesIterCache
}

func (rr *readRow) fill(series []*Series) {
	for _, ser := range series {
		rr.cols = append(rr.cols, &ser.col)
		rr.colReader = append(rr.colReader, rr.iteratorCache.Ensure(ser))
	}
}

// TODO return Row from this type
func (rr *readRow) Value() interface{} {
	r := Row{Index: Index(rr.index)}
	for c, sRead := range rr.colReader {
		r.Values = append(r.Values, Observation{col: rr.cols[c], value: sRead.Value()})
	}
	return r
}

// Sharing state across dependencies that need it
// TODO extract interfaces, potentially
// TODO if this were an ordered map, then it would be possible to make the assumption that dependencies
// have been advanced in extension columns, allowing us to cache values
type seriesIterCache map[*Series]iterator

func (c seriesIterCache) Advance() bool {
	for _, sRead := range c {
		if !sRead.Next() {
			return false
		}
	}
	return true
}

// nothing here is threadsafe!
func (c seriesIterCache) Ensure(ser *Series) iterator {
	if seriesRead, found := c[ser]; found {
		return seriesRead
	}
	// TODO prevent recursive loop using a visited set here ?
	seriesRead := ser.read(c)
	c[ser] = seriesRead
	return seriesRead
}

type tableIterator struct {
	readRow
}

func newTableIterator(series []*Series) *tableIterator {
	iter := &tableIterator{readRow{index: -1, iteratorCache: make(seriesIterCache)}}
	iter.fill(series)
	return iter
}

func (iter *tableIterator) Next() bool {
	// all series we depend on need to be advanced
	if !iter.iteratorCache.Advance() {
		return false
	}
	iter.index++
	return true
}
