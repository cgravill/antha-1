package data

type slicer struct {
	*seriesSlice
	iterator
	pos Index
}

func (iter *slicer) Next() bool {
	for iter.pos+1 < iter.start {
		if !iter.iterator.Next() {
			return false
		}
		iter.pos++
	}
	if !iter.iterator.Next() {
		// exhausted the underlying series
		return false
	}
	iter.pos++
	if iter.pos == iter.end {
		return false
	}
	return true
}

type seriesSlice struct {
	start, end Index
	wrapped    *Series
}

func newSeriesSlice(ser *Series, start, end Index) *seriesSlice {
	return &seriesSlice{start: start, end: end, wrapped: ser}
}

func (ss *seriesSlice) length() int {
	return int(ss.end - ss.start)
}

func (ss *seriesSlice) read(cache *seriesIterCache) iterator {
	return &slicer{
		seriesSlice: ss,
		pos:         -1,
		// TODO are we advancing series correctly here?
		iterator: ss.wrapped.read(cache),
	}
}

func (ss *seriesSlice) IsBounded() bool      { return ss.wrapped.meta.IsBounded() }
func (ss *seriesSlice) IsMaterialized() bool { return false }

func (ss *seriesSlice) ExactSize() int {
	length := ss.length()

	if length == 0 {
		return 0
	}
	if b, ok := ss.wrapped.meta.(Bounded); ok {
		w := b.ExactSize()
		if w == -1 {
			return -1
		}
		wrappedLen := w - int(ss.start)
		if wrappedLen < length {
			return wrappedLen
		}
		return length
	}
	return -1
}

func (ss *seriesSlice) MaxSize() int {
	length := ss.length()
	if b, ok := ss.wrapped.meta.(Bounded); ok {
		w := b.MaxSize() - int(ss.start)
		if w < length {
			return w
		}
	}
	return length
}

var _ Bounded = (*seriesSlice)(nil)
