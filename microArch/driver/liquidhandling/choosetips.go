package liquidhandling

import (
	"fmt"
	"github.com/pkg/errors"
	"sort"
	"strings"

	"github.com/Synthace/antha/antha/anthalib/wtype"
	"github.com/Synthace/antha/laboratory/effects/id"
)

type TipSource struct {
	DeckAddress string           // position of the tipbox from which to take tips
	WellAddress wtype.WellCoords // the well coordinate to use
}

type ChannelIndex int

func (ci ChannelIndex) String() string {
	return fmt.Sprintf("%d", ci)
}

type TipNotFoundError struct {
	Missing wtype.TipTypes // the tip types which were not found
}

func NewTipNotFoundError(missing ...wtype.TipType) *TipNotFoundError {
	return &TipNotFoundError{
		Missing: missing,
	}
}

func (tnf *TipNotFoundError) Error() string {
	return fmt.Sprintf("no tips found for type: %s", tnf.Missing)
}

// TipMask is a simplified tipbox representation which avoids the need to duplicate all tipboxes
// with every TipChooser call
type TipMask struct {
	Address string // the address of the deck slot on which this tipbox is placed
	Rows    int
	Cols    int
	TipType wtype.TipType
	Tips    map[wtype.WellCoords]bool // contains `true` for each coordinate which contains a useable tip, such that len(Tips) is the number of available tips
}

// NewTipMask create a tipmap to represent an existing tipbox
func NewTipMask(address string, tb *wtype.LHTipbox) *TipMask {
	tips := make(map[wtype.WellCoords]bool, tb.NRows()*tb.NCols())
	for x, col := range tb.Tips {
		for y, tip := range col {
			if tip != nil {
				// don't set false for tip == nil so len(tips) == number of tips
				tips[wtype.WellCoords{X: x, Y: y}] = true
			}
		}
	}
	return &TipMask{
		Address: address,
		Rows:    tb.NRows(),
		Cols:    tb.NCols(),
		TipType: tb.Tiptype.Type,
		Tips:    tips,
	}
}

func (tm *TipMask) AddressExists(wc wtype.WellCoords) bool {
	return wc.X >= 0 && wc.X < tm.Cols && wc.Y >= 0 && wc.Y < tm.Rows
}

func (tm *TipMask) NRows() int {
	return tm.Rows
}

func (tm *TipMask) NCols() int {
	return tm.Cols
}

func (tm *TipMask) GetChildByAddress(wc wtype.WellCoords) wtype.LHObject {
	if tm.Tips[wc] {
		return &wtype.LHTip{} // non-nil if tip exists
	}
	return nil
}

// CoordsToWellCoords required for Addressable
func (tm *TipMask) CoordsToWellCoords(*id.IDGenerator, wtype.Coordinates3D) (wtype.WellCoords, wtype.Coordinates3D) {
	return wtype.WellCoords{}, wtype.Coordinates3D{}
}

// WellCoordsToCoords required for Addressable, AddressIterator
func (tm *TipMask) WellCoordsToCoords(*id.IDGenerator, wtype.WellCoords, wtype.WellReference) (wtype.Coordinates3D, bool) {
	return wtype.Coordinates3D{}, false
}

// TipChooser a callback function which allows a device plugin to specify which tips to load
// given a copy of the preference-ordered list of available tipboxes, the head to load to, and which tip types should be loaded onto each channel, channelMap.
// If there is no error, the keys in the returned map should equal the keys in channelMap.
// If not enough tips were found, or the tips that were found couldn't be loaded a TipNotFoundError should be returned
type TipChooser func(tipboxes []*TipMask, head int, channelMap map[ChannelIndex]wtype.TipType) (map[ChannelIndex]TipSource, error)

// chooseTipsGilson TODO: this code should live in the instruction plugin and be provided as a callback once we no longer need to serialize
// Tip choice is quite constrained - tips must be contigious and are picked up from right-to-left by the left hand head and left-to-right by the right hand head.
// When more tips are required than are left in a column, extra tips are first picked up from the next column before returning to the original one.
func chooseTipsGilson(tipboxes []*TipMask, head int, channelMap map[ChannelIndex]wtype.TipType) (map[ChannelIndex]TipSource, error) {

	// first, some assertions

	// 1: there can be only one tiptype
	var tipType wtype.TipType
	tipTypes := make(map[wtype.TipType]bool, len(channelMap))
	for _, tt := range channelMap {
		tipTypes[tt] = true
		tipType = tt
	}
	if len(tipTypes) != 1 {
		s := make(wtype.TipTypes, 0, len(tipTypes))
		for tt := range tipTypes {
			s = append(s, tt)
		}
		return nil, errors.Errorf("Gilson device can only handle one tip type at a time: cannot load %v", s)
	}

	// 2: ChannelIndexes must be [0,7]
	invalid := make([]int, 0, len(channelMap))
	for ci := range channelMap {
		if ci < 0 || ci > 7 {
			invalid = append(invalid, int(ci))
		}
	}
	if len(invalid) > 0 {
		sort.Ints(invalid)
		return nil, errors.Errorf("Gilson device has only 8 channels: invalid channels %v", invalid)
	}

	// 3: channels must be contiguous, starting from 0
	minChannel := len(channelMap)
	maxChannel := 0
	for ci := range channelMap {
		if int(ci) < minChannel {
			minChannel = int(ci)
		}
		if int(ci) > maxChannel {
			maxChannel = int(ci)
		}
	}
	if minChannel != 0 || maxChannel+1 != len(channelMap) {
		return nil, errors.New("Gilson device can only load tips on contiguous channels")
	}

	// 4: head is 0 or 1
	if !(head == 0 || head == 1) {
		return nil, errors.Errorf("Unknown head %d", head)
	}

	// now let's figure out how to load these tips

	// let's try and find a tipbox with enough tips
	var box *TipMask
	for _, bx := range tipboxes {
		if bx.TipType == tipType && len(bx.Tips) >= len(channelMap) {
			box = bx
			break
		}
	}
	if box == nil {
		return nil, NewTipNotFoundError(tipType)
	}

	// loading direction for the heads -- nb. this depends on robot configuration, but likely head 0 is loaded on the left and head 1 on the right
	direction := []wtype.HorizontalDirection{wtype.RightToLeft, wtype.LeftToRight}
	it := wtype.NewAddressIterator(box, wtype.ColumnWise, wtype.BottomToTop, direction[head], false)
	coords := make([]wtype.WellCoords, 0, len(channelMap))
	for wc := it.Curr(); it.Valid() && len(coords) < len(channelMap); wc = it.Next() {
		if box.Tips[wc] {
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
			DeckAddress: box.Address,
			WellAddress: coords[len(coords)-1-i],
		}
	}

	return ret, nil
}

// chooseTipsHamilton TODO: this code should live in the instruction plugin and be provided as a callback once we no longer need to serialize
// The hamilton device has very few limitations for tip choice, so the behaviour here is to take tips in column order, back to front, left to
// right. This is done independently for each tip type, selecting tipboxes in preference order.
func chooseTipsHamilton(tipboxes []*TipMask, head int, channelMap map[ChannelIndex]wtype.TipType) (map[ChannelIndex]TipSource, error) {

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
	tipTypes := make([]wtype.TipType, 0, len(channelMap))
	for _, tt := range channelMap {
		if !tipTypeMap[tt] {
			tipTypeMap[tt] = true
			tipTypes = append(tipTypes, tt)
		}
	}

	// find the available tipboxes by tip type, in preference order
	tipboxesByType := make(map[wtype.TipType][]*TipMask, len(tipTypes))
	for _, tb := range tipboxes {
		tipboxesByType[tb.TipType] = append(tipboxesByType[tb.TipType], tb)
	}

	// initialise counters to keep track of which position of which box we're looking at for each type
	tipboxIndexByType := make(map[wtype.TipType]int, len(tipTypes))

	// moveToNextTip advance the iterator and tipbox index for this tip type to the next tip to be chosen
	getNextTip := func(tt wtype.TipType) (*TipMask, wtype.WellCoords, error) {
		tipboxes := tipboxesByType[tt]
		for tipboxIndexByType[tt] < len(tipboxes) {
			tb := tipboxes[tipboxIndexByType[tt]]
			it := wtype.NewAddressIterator(tb, wtype.ColumnWise, wtype.TopToBottom, wtype.LeftToRight, false)
			for wc := it.Curr(); it.Valid(); wc = it.Next() {
				if tb.Tips[wc] {
					tb.Tips[wc] = false // we need to remember that we took the tip
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
				DeckAddress: tb.Address,
				WellAddress: wc,
			}
		}
	}

	return ret, nil
}
