package liquidhandling

import (
	"fmt"
	"strings"
	"testing"

	"github.com/go-test/deep"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/laboratory/effects/id"
	"github.com/antha-lang/antha/workflow"
)

// TipChooserTipbox defines the state of a tipbox for testing, should be an array of 8 strings each of length 12
// where tct[i][j] == '1' means there's a tip at that position, any other value means there isn't
type TipChooserTestTipbox struct {
	TipType      string
	TipLocations [8]string
}

func (tct TipChooserTestTipbox) init(idGen *id.IDGenerator) *wtype.LHTipbox {
	shp := wtype.NewShape(wtype.CylinderShape, "mm", 7.3, 7.3, 51.2)
	w := wtype.NewLHWell(idGen, "ul", 250.0, 10.0, shp, wtype.FlatWellBottom, 7.3, 7.3, 51.2, 0.0, "mm")
	tiptype := wtype.NewLHTip(idGen, "me", tct.TipType, 0.5, 1000.0, "ul", false, shp, 44.7)
	tb := wtype.NewLHTipbox(idGen, 8, 12, wtype.Coordinates3D{127.76, 85.48, 120.0}, "me", "mytype", tiptype, w, 9.0, 9.0, 0.5, 0.5, 0.0)

	for i := 0; i < 8; i++ {
		for j := 0; j < 12; j++ {
			if tct.TipLocations[i][j] != '1' {
				tb.RemoveTip(wtype.WellCoords{X: j, Y: i})
			}
		}
	}

	return tb
}

type TipChooserTest struct {
	Name    string
	Chooser TipChooser // the function to test
	// Tipboxes define what tipboxes should be available for the test and what state they should be in.
	// Tipboxes are placed in positions with names "1","2","3", etc
	Tipboxes        []TipChooserTestTipbox
	Head            int
	ChannelMap      map[ChannelIndex]TipTypeName
	ExpectedError   string
	ExpectedSources map[ChannelIndex]TipSource
}

func (test *TipChooserTest) Run(t *testing.T) {
	t.Run(test.Name, test.run)
}

func (test *TipChooserTest) run(t *testing.T) {
	idGen := id.NewIDGenerator(fmt.Sprintf("test_%s", test.Name))

	// setup the tipboxes in a minimal lhproperties
	positions := make(map[string]*wtype.LHPosition, len(test.Tipboxes))
	names := make([]string, 0, len(test.Tipboxes))
	for i := range test.Tipboxes {
		name := fmt.Sprintf("%d", i+1)
		positions[name] = wtype.NewLHPosition(name, wtype.Coordinates3D{}, wtype.Coordinates2D{})
		names = append(names, name)
	}
	lhp := NewLHProperties(idGen, "testmodel", "testmnfr", LLLiquidHandler, DisposableTips, positions)
	lhp.Preferences = &workflow.LayoutOpt{
		Tipboxes: workflow.Addresses(names),
	}

	// initialise and add the tipboxes
	for _, tipbox := range test.Tipboxes {
		if err := lhp.AddTipBox(tipbox.init(idGen)); err != nil {
			t.Fatal(err)
		}
	}

	// now run the test
	sources, err := test.Chooser(lhp, test.Head, test.ChannelMap)
	if !test.expecting(err) {
		t.Fatalf("errors don't match:\ng: %s\ne: %s", err, test.ExpectedError)
	}

	if err == nil {
		if diffs := deep.Equal(sources, test.ExpectedSources); len(diffs) != 0 {
			t.Errorf("sources don't match:\n%s", strings.Join(diffs, "\n"))
		}

		// test that tips were removed from the locations
		for _, src := range sources {
			if lhp.Tipboxes[src.DeckAddress].HasTipAt(src.WellAddress) {
				t.Errorf("tip not removed from %s@%s", src.WellAddress.FormatA1(), src.DeckAddress)
			}
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

func TestGilsonTipChooser(t *testing.T) {
	TipChooserTests{
		{
			Name:    "head0-first tip in box",
			Chooser: chooseTipsGilson,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: "type1",
					TipLocations: [8]string{
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]TipTypeName{
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
					TipType: "type1",
					TipLocations: [8]string{
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
					},
				},
			},
			Head: 1,
			ChannelMap: map[ChannelIndex]TipTypeName{
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
					TipType: "type1",
					TipLocations: [8]string{
						"100000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]TipTypeName{
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
					TipType: "type1",
					TipLocations: [8]string{
						"000000000001",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
					},
				},
			},
			Head: 1,
			ChannelMap: map[ChannelIndex]TipTypeName{
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
					TipType: "type1",
					TipLocations: [8]string{
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]TipTypeName{
				0: "type1",
			},
			ExpectedError: "no tips found",
		},
		{
			Name:    "skip empty box",
			Chooser: chooseTipsGilson,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: "type1",
					TipLocations: [8]string{
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
						"000000000000",
					},
				},
				{
					TipType: "type1",
					TipLocations: [8]string{
						"111111000000",
						"111111000000",
						"111111000000",
						"111110000000",
						"111110000000",
						"111110000000",
						"111110000000",
						"111110000000",
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]TipTypeName{
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
					TipType: "type1",
					TipLocations: [8]string{
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
					},
				},
				{
					TipType: "type2",
					TipLocations: [8]string{
						"111111000000",
						"111111000000",
						"111111000000",
						"111110000000",
						"111110000000",
						"111110000000",
						"111110000000",
						"111110000000",
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]TipTypeName{
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
					TipType: "type1",
					TipLocations: [8]string{
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]TipTypeName{
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
					TipType: "type1",
					TipLocations: [8]string{
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111000",
						"111111111000",
						"111111111000",
						"111111111000",
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]TipTypeName{
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
					TipType: "type1",
					TipLocations: [8]string{
						"100000000000",
						"100000000000",
						"100000000000",
						"100000000000",
						"100000000000",
						"100000000000",
						"100000000000",
						"000000000000",
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]TipTypeName{
				0: "type1",
				1: "type1",
				2: "type1",
				3: "type1",
				4: "type1",
				5: "type1",
				6: "type1",
				7: "type1",
			},
			ExpectedError: "no tips found",
		},
		{
			Name:    "multi-types-fail",
			Chooser: chooseTipsGilson,
			Tipboxes: []TipChooserTestTipbox{
				{
					TipType: "type1",
					TipLocations: [8]string{
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
					},
				},
				{
					TipType: "type2",
					TipLocations: [8]string{
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
						"111111111100",
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]TipTypeName{
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
					TipType: "type1",
					TipLocations: [8]string{
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
						"111111111111",
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]TipTypeName{
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
					TipType: "type1",
					TipLocations: [8]string{
						"000000000111",
						"000000000111",
						"000000001111",
						"000000001111",
						"000000001111",
						"000000001111",
						"000000001111",
						"000000001111",
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]TipTypeName{
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
					TipType: "type1",
					TipLocations: [8]string{
						"000000000111",
						"000000000111",
						"000000001111",
						"000000001111",
						"000000001111",
						"000000001111",
						"000000001111",
						"000000001111",
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]TipTypeName{
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
					TipType: "type1",
					TipLocations: [8]string{
						"000000000111",
						"000000000111",
						"000000001111",
						"000000001111",
						"000000001111",
						"000000001111",
						"000000001111",
						"000000001111",
					},
				},
				{
					TipType: "type2",
					TipLocations: [8]string{
						"000000111111",
						"000000111111",
						"000000111111",
						"000001111111",
						"000001111111",
						"000001111111",
						"000001111111",
						"000001111111",
					},
				},
			},
			Head: 0,
			ChannelMap: map[ChannelIndex]TipTypeName{
				0: "type1",
				1: "type2",
				2: "type1",
				3: "type2",
				4: "type1",
				5: "type2",
				6: "type1",
				7: "type2",
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
