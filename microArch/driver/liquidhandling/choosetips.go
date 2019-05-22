package liquidhandling

import (
	"fmt"
	"github.com/pkg/errors"
	"sort"
	"strings"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
)

type TipSource struct {
	TipboxID    string           // the ID of the tipbox from which to take tips
	WellAddress wtype.WellCoords // the well coordinate to use
}

type ChannelIndex int

func (ci ChannelIndex) String() string {
	return fmt.Sprintf("%d", ci)
}

type TipNotFoundError struct {
	Missing []wtype.TipType // the tip types which were not found
}

func NewTipNotFoundError(missing ...wtype.TipType) *TipNotFoundError {
	return &TipNotFoundError{
		Missing: missing,
	}
}

func (tnf *TipNotFoundError) Error() string {
	s := make([]string, len(tnf.Missing))
	for i, m := range tnf.Missing {
		s[i] = string(m)
	}
	sort.Strings(s)
	return fmt.Sprintf("no tips found for type: %s", strings.Join(s, ", "))
}

// TipChooser a callback function which allows a device plugin to specify which tips to load
// given a copy of the preference-ordered list of available tipboxes, the head to load to, and which tip types should be loaded onto each channel, channelMap.
// If there is no error, the keys in the returned map should equal the keys in channelMap.
// If not enough tips were found, or the tips that were found couldn't be loaded a TipNotFoundError should be returned
type TipChooser func(tipboxes []*wtype.LHTipbox, head int, channelMap map[ChannelIndex]wtype.TipType) (map[ChannelIndex]TipSource, error)

// chooseTipsGilson TODO: this code should live in the instruction plugin and be provided as a callback once we no longer need to serialize
// Tip choice is quite constrained - tips must be contigious and are picked up from right-to-left by the left hand head and left-to-right by the right hand head.
// When more tips are required than are left in a column, extra tips are first picked up from the next column before returning to the original one.
func chooseTipsGilson(tipboxes []*wtype.LHTipbox, head int, channelMap map[ChannelIndex]wtype.TipType) (map[ChannelIndex]TipSource, error) {

	// first, some assertions

	// 1: there can be only one tiptype
	var tipType wtype.TipType
	tipTypes := make(map[wtype.TipType]bool, len(channelMap))
	for _, tt := range channelMap {
		tipTypes[tt] = true
		tipType = tt
	}
	if len(tipTypes) != 1 {
		s := make([]string, 0, len(tipTypes))
		for tt := range tipTypes {
			s = append(s, string(tt))
		}
		sort.Strings(s)
		return nil, errors.Errorf("Gilson device can only handle one tip type at a time: cannot load %s", strings.Join(s, ", "))
	}

	// 2: ChannelIndexes must be [0,7]
	invalid := make([]string, 0, len(channelMap))
	for ci := range channelMap {
		if ci < 0 || ci > 7 {
			invalid = append(invalid, fmt.Sprintf("%d", ci))
		}
	}
	if len(invalid) > 0 {
		sort.Strings(invalid)
		return nil, errors.Errorf("Gilson device has only 8 channels: invalid channels %s", strings.Join(invalid, ", "))
	}

	// 3: channels must be contiguous, starting from 0
	missing := make([]string, 0, len(channelMap))
	for i := 0; i < len(channelMap); i++ {
		if _, ok := channelMap[ChannelIndex(i)]; !ok {
			missing = append(missing, fmt.Sprintf("%d", i))
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return nil, errors.Errorf("Gilson device can only load tips on contiguous channels: skipping channels %s", strings.Join(missing, ", "))
	}

	// 4: head is 0 or 1
	if !(head == 0 || head == 1) {
		return nil, errors.Errorf("Unknown head %d", head)
	}

	// not let's figure out how to load these tips

	// let's try and find a tipbox with enough tips
	var box *wtype.LHTipbox
	for _, bx := range tipboxes {
		if wtype.TipType(bx.Tiptype.Type) == tipType && bx.N_clean_tips() >= len(channelMap) {
			box = bx
		}
	}
	if box == nil {
		return nil, NewTipNotFoundError(tipType)
	}

	// loading direction for the heads -- nb. this depends on robot configuration, but likely head 0 is loaded on the left and head 1 on the right
	direction := map[int]wtype.HorizontalDirection{
		0: wtype.RightToLeft,
		1: wtype.LeftToRight,
	}
	it := wtype.NewAddressIterator(box, wtype.ColumnWise, wtype.BottomToTop, direction[head], false)
	coords := make([]wtype.WellCoords, 0, len(channelMap))
	for wc := it.Curr(); it.Valid() && len(coords) < len(channelMap); wc = it.Next() {
		if box.HasTipAt(wc) {
			box.RemoveTip(wc)
			coords = append(coords, wc)
		}
	}
	if len(coords) != len(channelMap) {
		// shouldn't happen because we checked N_clean_tips earlier
		panic("failed to find enough tips")
	}

	// tips are picked up in reverse order
	ret := make(map[ChannelIndex]TipSource, len(channelMap))
	for i := 0; i < len(channelMap); i++ {
		ret[ChannelIndex(i)] = TipSource{
			TipboxID:    box.ID,
			WellAddress: coords[len(coords)-1-i],
		}
	}

	return ret, nil
}

// chooseTipsHamilton TODO: this code should live in the instruction plugin and be provided as a callback once we no longer need to serialize
// The hamilton device has very few limitations for tip choice, so the behaviour here is to take tips in column order, back to front, left to
// right. This is done independently for each tip type, selecting tipboxes in preference order.
func chooseTipsHamilton(tipboxes []*wtype.LHTipbox, head int, channelMap map[ChannelIndex]wtype.TipType) (map[ChannelIndex]TipSource, error) {

	// some assertions

	// 1. we only support head = 0
	if head != 0 {
		return nil, errors.Errorf("head %d not supported", head)
	}

	// 2. channelIndexes must be [0,7]
	invalid := make([]string, 0, len(channelMap))
	for ci := range channelMap {
		if ci < 0 || ci > 7 {
			invalid = append(invalid, fmt.Sprintf("%d", ci))
		}
	}
	if len(invalid) > 0 {
		sort.Strings(invalid)
		return nil, errors.Errorf("invalid channel indexes: %s", strings.Join(invalid, ", "))
	}

	// get unique tiptypes
	tipTypeMap := make(map[wtype.TipType]bool, len(channelMap))
	for _, tt := range channelMap {
		tipTypeMap[tt] = true
	}
	tipTypes := make([]wtype.TipType, 0, len(tipTypeMap))
	for tt := range tipTypeMap {
		tipTypes = append(tipTypes, tt)
	}

	// find the available tipboxes by tip type, in preference order
	tipboxesByType := make(map[wtype.TipType][]*wtype.LHTipbox, len(tipTypes))
	for _, tb := range tipboxes {
		tipboxesByType[wtype.TipType(tb.Tiptype.Type)] = append(tipboxesByType[wtype.TipType(tb.Tiptype.Type)], tb)
	}

	// initialise counters to keep track of which position of which box we're looking at for each type
	tipboxIndexByType := make(map[wtype.TipType]int, len(tipTypes))

	// moveToNextTip advance the iterator and tipbox index for this tip type to the next tip to be chosen
	getNextTip := func(tt wtype.TipType) (*wtype.LHTipbox, wtype.WellCoords, error) {
		tipboxes := tipboxesByType[tt]
		for tipboxIndexByType[tt] < len(tipboxes) {
			tb := tipboxes[tipboxIndexByType[tt]]
			it := wtype.NewAddressIterator(tb, wtype.ColumnWise, wtype.TopToBottom, wtype.LeftToRight, false)
			for wc := it.Curr(); it.Valid(); wc = it.Next() {
				if tb.HasTipAt(wc) {
					tb.RemoveTip(wc)
					return tb, wc, nil
				}
			}

			// this tipbox is empty, move to the next
			tipboxIndexByType[tt] += 1
		}
		// we ran out of tipboxes for this type
		return nil, wtype.WellCoords{}, NewTipNotFoundError(tt)
	}

	// need to go in channel order
	channels := make([]int, 0, len(channelMap))
	for ci := range channelMap {
		channels = append(channels, int(ci))
	}
	sort.Ints(channels)

	// now find the tips
	ret := make(map[ChannelIndex]TipSource, len(channelMap))
	for _, ci := range channels {
		if tb, wc, err := getNextTip(channelMap[ChannelIndex(ci)]); err != nil {
			return ret, err
		} else {
			ret[ChannelIndex(ci)] = TipSource{
				TipboxID:    tb.ID,
				WellAddress: wc,
			}
		}
	}

	return ret, nil
}
