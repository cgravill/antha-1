package liquidhandling

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-test/deep"

	"github.com/Synthace/antha/antha/anthalib/wtype"
)

// TipChooserTipbox defines the state of a tipbox for testing, should be an array of 8 strings each of length 12
// where tct[i][j] == '1' means there's a tip at that position, any other value means there isn't
type TipChooserTestTipbox struct {
	TipType      wtype.TipType
	TipLocations [8]string
}

func (tct TipChooserTestTipbox) Mask(pos string) *TipMask {

	tips := make(map[wtype.WellCoords]bool, 96)
	for i := 0; i < 8; i++ {
		for j := 0; j < 12; j++ {
			if tct.TipLocations[i][j] == '1' {
				tips[wtype.WellCoords{X: j, Y: i}] = true
			}
		}
	}

	return &TipMask{
		Address: pos,
		Rows:    8,
		Cols:    12,
		TipType: tct.TipType,
		Tips:    tips,
	}
}

type TipChooserTest struct {
	Name    string
	Chooser TipChooser // the function to test
	// Tipboxes define what tipboxes should be available for the test and what state they should be in.
	// Tipbox IDs are replaced with "1","2","3", etc for testing
	Tipboxes        []TipChooserTestTipbox
	Head            int
	ChannelMap      map[ChannelIndex]wtype.TipType
	ExpectedError   string
	ExpectedSources map[ChannelIndex]TipSource
}

func (test *TipChooserTest) Run(t *testing.T) {
	t.Run(test.Name, test.run)
}

func (test *TipChooserTest) run(t *testing.T) {

	// setup the tipboxes
	tipboxes := make([]*TipMask, len(test.Tipboxes))
	for i, tipbox := range test.Tipboxes {
		tipboxes[i] = tipbox.Mask(fmt.Sprintf("%d", i+1))
	}

	// now run the test
	sources, err := test.Chooser(tipboxes, test.Head, test.ChannelMap)
	if !test.expecting(err) {
		t.Fatalf("errors don't match:\ng: %s\ne: %s", err, test.ExpectedError)
	}

	if err == nil {
		if diffs := deep.Equal(sources, test.ExpectedSources); len(diffs) != 0 {
			t.Errorf("sources don't match:\n%s", strings.Join(diffs, "\n"))
		}
	}

}

func (test *TipChooserTest) expecting(err error) bool {
	if (err == nil) != (test.ExpectedError == "") {
		return false
	}
	if err != nil {
		return test.ExpectedError == err.Error()
	}
	return true
}

type TipChooserTests []TipChooserTest

func (tests TipChooserTests) Run(t *testing.T) {
	for _, test := range tests {
		test.Run(t)
	}
}

var type1 = wtype.TipType("type1")
var type2 = wtype.TipType("type2")

func TestGilsonTipChooser(t *testing.T) {
	TipChooserTests{
		{
			Name:    "head0-first tip in box",
			Chooser: chooseTipsGilson,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"111111111111", // A
						"111111111111", // B
						"111111111111", // C
						"111111111111", // D
						"111111111111", // E
						"111111111111", // F
						"111111111111", // G
						"111111111111", // H
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: "type1",
			},
			ExpectedSources: map[ChannelIndex]TipSource{
				0: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("H12"),
				},
			},
		},
		{
			Name:    "head1-first tip in box",
			Chooser: chooseTipsGilson,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"111111111111", // A
						"111111111111", // B
						"111111111111", // C
						"111111111111", // D
						"111111111111", // E
						"111111111111", // F
						"111111111111", // G
						"111111111111", // H
					},
				},
			},
			Head: 1,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: "type1",
			},
			ExpectedSources: map[ChannelIndex]TipSource{
				0: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("H1"),
				},
			},
		},
		{
			Name:    "head0-last tip in box",
			Chooser: chooseTipsGilson,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"100000000000", // A
						"000000000000", // B
						"000000000000", // C
						"000000000000", // D
						"000000000000", // E
						"000000000000", // F
						"000000000000", // G
						"000000000000", // H
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: "type1",
			},
			ExpectedSources: map[ChannelIndex]TipSource{
				0: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("A1"),
				},
			},
		},
		{
			Name:    "head1-last tip in box",
			Chooser: chooseTipsGilson,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"000000000001", // A
						"000000000000", // B
						"000000000000", // C
						"000000000000", // D
						"000000000000", // E
						"000000000000", // F
						"000000000000", // G
						"000000000000", // H
					},
				},
			},
			Head: 1,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: "type1",
			},
			ExpectedSources: map[ChannelIndex]TipSource{
				0: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("A12"),
				},
			},
		},
		{
			Name:    "no tips",
			Chooser: chooseTipsGilson,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"000000000000", // A
						"000000000000", // B
						"000000000000", // C
						"000000000000", // D
						"000000000000", // E
						"000000000000", // F
						"000000000000", // G
						"000000000000", // H
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: "type1",
			},
			ExpectedError: "no tips found for type: type1",
		},
		{
			Name:    "skip empty box",
			Chooser: chooseTipsGilson,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"000000000000", // A
						"000000000000", // B
						"000000000000", // C
						"000000000000", // D
						"000000000000", // E
						"000000000000", // F
						"000000000000", // G
						"000000000000", // H
					},
				},
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"111111000000", // A
						"111111000000", // B
						"111111000000", // C
						"111110000000", // D
						"111110000000", // E
						"111110000000", // F
						"111110000000", // G
						"111110000000", // H
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: "type1",
			},
			ExpectedSources: map[ChannelIndex]TipSource{
				0: {
					DeckAddress: "2",
					WellAddress: wtype.MakeWellCoords("C6"),
				},
			},
		},
		{
			Name:    "ignore wrong type",
			Chooser: chooseTipsGilson,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"111111111111", // A
						"111111111111", // B
						"111111111111", // C
						"111111111111", // D
						"111111111111", // E
						"111111111111", // F
						"111111111111", // G
						"111111111111", // H
					},
				},
				{
					TipType: type2,
					TipLocations: [8]string{
						//        111
						//23456789012
						"111111000000", // A
						"111111000000", // B
						"111111000000", // C
						"111110000000", // D
						"111110000000", // E
						"111110000000", // F
						"111110000000", // G
						"111110000000", // H
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: "type2",
			},
			ExpectedSources: map[ChannelIndex]TipSource{
				0: {
					DeckAddress: "2",
					WellAddress: wtype.MakeWellCoords("C6"),
				},
			},
		},
		{
			Name:    "head0-multiCol",
			Chooser: chooseTipsGilson,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"111111111100", // A
						"111111111100", // B
						"111111111100", // C
						"111111111100", // D
						"111111111100", // E
						"111111111100", // F
						"111111111100", // G
						"111111111100", // H
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: "type1",
				1: "type1",
				2: "type1",
				3: "type1",
				4: "type1",
				5: "type1",
				6: "type1",
				7: "type1",
			},
			ExpectedSources: map[ChannelIndex]TipSource{
				0: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("A10"),
				},
				1: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("B10"),
				},
				2: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("C10"),
				},
				3: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("D10"),
				},
				4: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("E10"),
				},
				5: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("F10"),
				},
				6: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("G10"),
				},
				7: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("H10"),
				},
			},
		},
		{
			Name:    "head0-multi-Chunked",
			Chooser: chooseTipsGilson,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"111111111100", // A
						"111111111100", // B
						"111111111100", // C
						"111111111100", // D
						"111111111000", // E
						"111111111000", // F
						"111111111000", // G
						"111111111000", // H
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: "type1",
				1: "type1",
				2: "type1",
				3: "type1",
				4: "type1",
				5: "type1",
				6: "type1",
				7: "type1",
			},
			ExpectedSources: map[ChannelIndex]TipSource{
				0: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("E9"),
				},
				1: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("F9"),
				},
				2: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("G9"),
				},
				3: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("H9"),
				},
				4: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("A10"),
				},
				5: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("B10"),
				},
				6: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("C10"),
				},
				7: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("D10"),
				},
			},
		},
		{
			Name:    "head0-multi-notips",
			Chooser: chooseTipsGilson,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"100000000000", // A
						"100000000000", // B
						"100000000000", // C
						"100000000000", // D
						"100000000000", // E
						"100000000000", // F
						"100000000000", // G
						"000000000000", // H
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: "type1",
				1: "type1",
				2: "type1",
				3: "type1",
				4: "type1",
				5: "type1",
				6: "type1",
				7: "type1",
			},
			ExpectedError: "no tips found for type: type1",
		},
		{
			Name:    "multi-types-fail",
			Chooser: chooseTipsGilson,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"111111111100", // A
						"111111111100", // B
						"111111111100", // C
						"111111111100", // D
						"111111111100", // E
						"111111111100", // F
						"111111111100", // G
						"111111111100", // H
					},
				},
				{
					TipType: type2,
					TipLocations: [8]string{
						//        111
						//23456789012
						"111111111100", // A
						"111111111100", // B
						"111111111100", // C
						"111111111100", // D
						"111111111100", // E
						"111111111100", // F
						"111111111100", // G
						"111111111100", // H
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: "type1",
				1: "type1",
				2: "type1",
				3: "type1",
				4: "type2",
				5: "type2",
				6: "type2",
				7: "type2",
			},
			ExpectedError: "Gilson device can only handle one tip type at a time: cannot load type1, type2",
		},
	}.Run(t)
}

func TestHamiltonTipChooser(t *testing.T) {
	TipChooserTests{
		{
			Name:    "first tip in box",
			Chooser: chooseTipsHamilton,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"111111111111", // A
						"111111111111", // B
						"111111111111", // C
						"111111111111", // D
						"111111111111", // E
						"111111111111", // F
						"111111111111", // G
						"111111111111", // H
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: "type1",
			},
			ExpectedSources: map[ChannelIndex]TipSource{
				0: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("A1"),
				},
			},
		},
		{
			Name:    "random tip in box",
			Chooser: chooseTipsHamilton,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"000000000111", // A
						"000000000111", // B
						"000000001111", // C
						"000000001111", // D
						"000000001111", // E
						"000000001111", // F
						"000000001111", // G
						"000000001111", // H
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: "type1",
			},
			ExpectedSources: map[ChannelIndex]TipSource{
				0: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("C9"),
				},
			},
		},
		{
			Name:    "multi",
			Chooser: chooseTipsHamilton,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"000000000111", // A
						"000000000111", // B
						"000000001111", // C
						"000000001111", // D
						"000000001111", // E
						"000000001111", // F
						"000000001111", // G
						"000000001111", // H
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: "type1",
				2: "type1",
				4: "type1",
				6: "type1",
			},
			ExpectedSources: map[ChannelIndex]TipSource{
				0: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("C9"),
				},
				2: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("D9"),
				},
				4: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("E9"),
				},
				6: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("F9"),
				},
			},
		},
		{
			Name:    "multi-types",
			Chooser: chooseTipsHamilton,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: type1,
					TipLocations: [8]string{
						//        111
						//23456789012
						"000000000111", // A
						"000000000111", // B
						"000000001111", // C
						"000000001111", // D
						"000000001111", // E
						"000000001111", // F
						"000000001111", // G
						"000000001111", // H
					},
				},
				{
					TipType: type2,
					TipLocations: [8]string{
						//        111
						//23456789012
						"000000111111", // A
						"000000111111", // B
						"000000111111", // C
						"000001111111", // D
						"000001111111", // E
						"000001111111", // F
						"000001111111", // G
						"000001111111", // H
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]wtype.TipType{
				0: type1,
				1: type2,
				2: type1,
				3: type2,
				4: type1,
				5: type2,
				6: type1,
				7: type2,
			},
			ExpectedSources: map[ChannelIndex]TipSource{
				0: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("C9"),
				},
				1: {
					DeckAddress: "2",
					WellAddress: wtype.MakeWellCoords("D6"),
				},
				2: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("D9"),
				},
				3: {
					DeckAddress: "2",
					WellAddress: wtype.MakeWellCoords("E6"),
				},
				4: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("E9"),
				},
				5: {
					DeckAddress: "2",
					WellAddress: wtype.MakeWellCoords("F6"),
				},
				6: {
					DeckAddress: "1",
					WellAddress: wtype.MakeWellCoords("F9"),
				},
				7: {
					DeckAddress: "2",
					WellAddress: wtype.MakeWellCoords("G6"),
				},
			},
		},
	}.Run(t)
}
