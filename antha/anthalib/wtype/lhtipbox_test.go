package wtype

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/antha-lang/antha/laboratory/effects/id"
)

func makeTipboxForTest(idGen *id.IDGenerator) *LHTipbox {
	shp := NewShape(CylinderShape, "mm", 7.3, 7.3, 51.2)
	w := NewLHWell(idGen, "ul", 250.0, 10.0, shp, FlatWellBottom, 7.3, 7.3, 51.2, 0.0, "mm")
	tiptype := makeTipForTest(idGen)
	tb := NewLHTipbox(idGen, 8, 12, Coordinates3D{127.76, 85.48, 120.0}, "me", "mytype", tiptype, w, 9.0, 9.0, 0.5, 0.5, 0.0)
	return tb
}

func TestTipboxWellCoordsToCoords(t *testing.T) {
	idGen := id.NewIDGenerator(t.Name())

	tb := makeTipboxForTest(idGen)

	pos, ok := tb.WellCoordsToCoords(idGen, MakeWellCoords("A1"), BottomReference)
	if !ok {
		t.Fatal("well A1 doesn't exist!")
	}

	xExpected := tb.TipXStart
	yExpected := tb.TipYStart

	if pos.X != xExpected || pos.Y != yExpected {
		t.Errorf("position was wrong: expected (%f, %f) got (%f, %f)", xExpected, yExpected, pos.X, pos.Y)
	}

}

func TestTipboxCoordsToWellCoords(t *testing.T) {
	idGen := id.NewIDGenerator(t.Name())

	tb := makeTipboxForTest(idGen)

	pos := Coordinates3D{
		X: tb.TipXStart + 0.75*tb.TipXOffset,
		Y: tb.TipYStart + 0.75*tb.TipYOffset,
	}

	wc, delta := tb.CoordsToWellCoords(idGen, pos)

	if e, g := "B2", wc.FormatA1(); e != g {
		t.Errorf("Wrong well coordinates: expected %s, got %s", e, g)
	}

	eDelta := -0.25 * tb.TipXOffset
	if delta.X != eDelta || delta.Y != eDelta {
		t.Errorf("Delta incorrect: expected (%f, %f), got (%f, %f)", eDelta, eDelta, delta.X, delta.Y)
	}

}

func TestTipboxGetWellBounds(t *testing.T) {
	idGen := id.NewIDGenerator(t.Name())

	tb := makeTipboxForTest(idGen)

	eStart := Coordinates3D{
		X: 0.5 - 0.5*7.3,
		Y: 0.5 - 0.5*7.3,
		Z: 0.0,
	}
	eSize := Coordinates3D{
		X: 9.0*11 + 7.3,
		Y: 9.0*7 + 7.3,
		Z: 51.2,
	}
	eBounds := NewBBox(eStart, eSize)
	bounds := tb.GetTipBounds()

	if e, g := eBounds.String(), bounds.String(); e != g {
		t.Errorf("GetWellBounds incorrect: expected %v, got %v", eBounds, bounds)
	}
}

func TestTipboxSerialization(t *testing.T) {
	idGen := id.NewIDGenerator(t.Name())

	removed := makeTipboxForTest(idGen)
	toRemove := MakeWellCoordsArray([]string{"A1", "B2", "H5"})
	for _, wc := range toRemove {
		removed.RemoveTip(wc)
	}

	if e, g := (removed.NCols()*removed.NRows() - len(toRemove)), removed.N_clean_tips(); e != g {
		t.Fatalf("LHTipbox.RemoveTip didn't work: expected %d tips remaining, got %d", e, g)
	}

	tipboxes := []*LHTipbox{
		makeTipboxForTest(idGen),
	}

	for _, before := range tipboxes {

		var after LHTipbox
		if bs, err := json.Marshal(before); err != nil {
			t.Fatal(err)
		} else if err := json.Unmarshal(bs, &after); err != nil {
			t.Fatal(err)
		}

		for _, row := range after.Tips {
			for _, tip := range row {
				if tip.parent != &after {
					t.Fatal("parent not set correctly")
				}
			}
		}

		if !reflect.DeepEqual(before, &after) {
			t.Errorf("serialization changed the tipbox:\nbefore: %+v\n after: %+v", before, &after)
		}
	}
}
