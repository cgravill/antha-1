// liquidhandling/lhtypes.Go: Part of the Antha language
// Copyright (C) 2014 the Antha authors. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
//
// For more information relating to the software or licensing issues please
// contact license@antha-lang.Org or write to the Antha team c/o
// Synthace Ltd. The London Bioscience Innovation Centre
// 2 Royal College St, London NW1 0NH UK

package wtype

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"

	"github.com/antha-lang/antha/laboratory/effects/id"
)

/* tip box */

type LHTipbox struct {
	ID         string
	Boxname    string
	Type       string
	Mnfr       string
	Nrows      int
	Ncols      int
	Height     float64
	Tiptype    *LHTip
	AsWell     *LHWell
	NTips      int
	Tips       [][]*LHTip // Tips[row][col]
	TipXOffset float64
	TipYOffset float64
	TipXStart  float64
	TipYStart  float64
	TipZStart  float64

	Bounds BBox
	parent LHObject `gotopb:"-"`
}

func NewLHTipbox(idGen *id.IDGenerator, nrows, ncols int, size Coordinates3D, manufacturer, boxtype string, tiptype *LHTip, well *LHWell, tipxoffset, tipyoffset, tipxstart, tipystart, tipzstart float64) *LHTipbox {
	var tipbox LHTipbox
	tipbox.ID = idGen.NextID()
	tipbox.Type = boxtype
	tipbox.Boxname = fmt.Sprintf("%s_%s", boxtype, tipbox.ID[1:len(tipbox.ID)-2])
	tipbox.Mnfr = manufacturer
	tipbox.Nrows = nrows
	tipbox.Ncols = ncols
	tipbox.Tips = make([][]*LHTip, ncols)
	tipbox.NTips = tipbox.Nrows * tipbox.Ncols
	tipbox.Bounds.SetSize(size)
	tipbox.Tiptype = tiptype
	tipbox.AsWell = well
	for i := 0; i < ncols; i++ {
		tipbox.Tips[i] = make([]*LHTip, nrows)
	}
	tipbox.TipXOffset = tipxoffset
	tipbox.TipYOffset = tipyoffset
	tipbox.TipXStart = tipxstart
	tipbox.TipYStart = tipystart
	tipbox.TipZStart = tipzstart

	return initialize_tips(idGen, &tipbox, tiptype)
}

func (tb LHTipbox) GetID() string {
	return tb.ID
}

func (tb LHTipbox) Output() string {
	s := ""
	for j := 0; j < tb.NRows(); j++ {
		for i := 0; i < tb.NCols(); i++ {
			if tb.Tips[i][j] == nil {
				s += "."
			} else if tb.Tips[i][j].Dirty {
				s += "*"
			} else {
				s += "o"
			}
		}
		s += "\n"
	}

	return s
}

func (tb LHTipbox) String() string {
	return fmt.Sprintf(
		`LHTipbox {
ID        : %s,
Boxname   : %s,
Type      : %s,
Mnfr      : %s,
Nrows     : %d,
Ncols     : %d,
Width     : %f,
Length    : %f,
Height    : %f,
Tiptype   : %p,
AsWell    : %v,
NTips     : %d,
Tips      : %p,
TipXOffset: %f,
TipYOffset: %f,
TipXStart : %f,
TipYStart : %f,
TipZStart : %f,
}`,
		tb.ID,
		tb.Boxname,
		tb.Type,
		tb.Mnfr,
		tb.Nrows,
		tb.Ncols,
		tb.Bounds.GetSize().X,
		tb.Bounds.GetSize().Y,
		tb.Bounds.GetSize().Z,
		tb.Tiptype,
		tb.AsWell,
		tb.NTips,
		tb.Tips,
		tb.TipXOffset,
		tb.TipYOffset,
		tb.TipXStart,
		tb.TipYStart,
		tb.TipZStart,
	)
}

func (tb *LHTipbox) Dup(idGen *id.IDGenerator) *LHTipbox {
	return tb.dup(idGen, false)
}
func (tb *LHTipbox) DupKeepIDs(idGen *id.IDGenerator) *LHTipbox {
	return tb.dup(idGen, true)
}

func (tb *LHTipbox) dup(idGen *id.IDGenerator, keepIDs bool) *LHTipbox {
	tb2 := NewLHTipbox(idGen, tb.Nrows, tb.Ncols, tb.Bounds.GetSize(), tb.Mnfr, tb.Type, tb.Tiptype, tb.AsWell, tb.TipXOffset, tb.TipYOffset, tb.TipXStart, tb.TipYStart, tb.TipZStart)
	tb2.Bounds.Position = tb.Bounds.GetPosition()

	if keepIDs {
		tb2.ID = tb.ID
		//boxname contains the ID
		tb2.Boxname = tb.Boxname
	}

	for i := 0; i < len(tb.Tips); i++ {
		for j := 0; j < len(tb.Tips[i]); j++ {
			t := tb.Tips[i][j]
			if t == nil {
				tb2.Tips[i][j] = nil
			} else {
				if keepIDs {
					tb2.Tips[i][j] = t.DupKeepID(idGen)
				} else {
					tb2.Tips[i][j] = t.Dup(idGen)
				}
				tb2.Tips[i][j].SetParent(tb2) //nolint - tb2 is certainly an lhtipbox
			}
		}
	}

	return tb2
}

// @implement named

func (tb *LHTipbox) GetName() string {
	if tb == nil {
		return "<nil>"
	}
	return tb.Boxname
}

func (tb *LHTipbox) GetType() string {
	if tb == nil {
		return "<nil>"
	}
	return tb.Type
}

func (self *LHTipbox) GetClass() string {
	return "tipbox"
}

func (tb *LHTipbox) N_clean_tips() int {
	c := 0
	for j := 0; j < tb.Nrows; j++ {
		for i := 0; i < tb.Ncols; i++ {
			if tb.Tips[i][j] != nil && !tb.Tips[i][j].Dirty {
				c += 1
			}
		}
	}
	return c
}

//HasEnoughTips returns true if the tipbox has at least requested tips
//equivalent to tb.N_clean_tips() > requested
func (tb *LHTipbox) HasEnoughTips(requested int) bool {
	c := 0
	for _, tiprow := range tb.Tips {
		for _, tip := range tiprow {
			if tip != nil && !tip.Dirty {
				c += 1
				if c >= requested {
					return true
				}
			}
		}
	}
	return c >= requested
}

//##############################################
//@implement LHObject
//##############################################

func (self *LHTipbox) GetPosition() Coordinates3D {
	if self.parent != nil {
		return self.parent.GetPosition().Add(self.Bounds.GetPosition())
	}
	return self.Bounds.GetPosition()
}

func (self *LHTipbox) GetSize() Coordinates3D {
	return self.Bounds.GetSize()
}

func (self *LHTipbox) GetTipBounds() BBox {
	tipSize := self.Tiptype.GetSize()

	pos := self.Bounds.GetPosition().Add(Coordinates3D{
		X: self.TipXStart - 0.5*tipSize.X,
		Y: self.TipYStart - 0.5*tipSize.Y,
		Z: self.TipZStart})

	size := Coordinates3D{
		X: self.TipXOffset*float64(self.NCols()-1) + tipSize.X,
		Y: self.TipYOffset*float64(self.NRows()-1) + tipSize.Y,
		Z: tipSize.Z,
	}
	return BBox{pos, size}
}

func (self *LHTipbox) GetBoxIntersections(box BBox) []LHObject {
	//relative box
	relBox := NewBBox(box.GetPosition().Subtract(OriginOf(self)), box.GetSize())
	ret := []LHObject{}
	if self.Bounds.IntersectsBox(*relBox) {
		ret = append(ret, self)
	}

	//if it's possible the this box might intersect with some tips
	if self.GetTipBounds().IntersectsBox(*relBox) {
		for _, tiprow := range self.Tips {
			for _, tip := range tiprow {
				if tip != nil {
					c := tip.GetBoxIntersections(box)
					if c != nil {
						ret = append(ret, c...)
					}
				}
			}
		}
	}
	return ret
}

func trimToMask(wells []string, mask []bool) []string {
	if len(mask) >= len(wells) {
		return wells
	}
	ret := make([]string, len(mask))
	s := false
	x := 0
	for i := 0; i < len(wells); i++ {
		if wells[i] != "" && !s {
			s = true
		}

		if s {
			ret[x] = wells[i]
			x += 1
		}

		if x == len(mask) {
			break
		}
	}
	return ret
}

func (self *LHTipbox) GetPointIntersections(point Coordinates3D) []LHObject {
	//relative point
	relPoint := point.Subtract(OriginOf(self))
	ret := []LHObject{}
	if self.Bounds.IntersectsPoint(relPoint) {
		ret = append(ret, self)
	}

	//if it's possible the this point might intersect with some tips
	if self.GetTipBounds().IntersectsPoint(relPoint) {
		for _, tiprow := range self.Tips {
			for _, tip := range tiprow {
				ret = append(ret, tip.GetPointIntersections(point)...)
			}
		}
	}
	return ret
}

func (self *LHTipbox) SetOffset(o Coordinates3D) error {
	self.Bounds.SetPosition(o)
	return nil
}

func (self *LHTipbox) SetParent(p LHObject) error {
	self.parent = p
	return nil
}

//@implement LHObject
func (self *LHTipbox) ClearParent() {
	self.parent = nil
}

func (self *LHTipbox) GetParent() LHObject {
	return self.parent
}

//Duplicate copies an LHObject
func (self *LHTipbox) Duplicate(idGen *id.IDGenerator, keepIDs bool) LHObject {
	return self.dup(idGen, keepIDs)
}

//DimensionsString returns a string description of the position and size of the object and its children.
//useful for debugging
func (self *LHTipbox) DimensionsString(idGen *id.IDGenerator) string {
	ret := make([]string, 0, 1+self.NRows()*self.NCols())
	ret = append(ret, fmt.Sprintf("Tipbox \"%s\" at %v+%v, with %dx%d tips bounded by %v",
		self.GetName(), self.GetPosition(), self.GetSize(), self.NCols(), self.NRows(), self.GetTipBounds()))

	for _, tiprow := range self.Tips {
		for _, tip := range tiprow {
			ret = append(ret, "\t"+tip.DimensionsString(idGen))
		}
	}

	return strings.Join(ret, "\n")
}

//##############################################
//@implement Addressable
//##############################################

func (tb *LHTipbox) AddressExists(c WellCoords) bool {
	return c.X >= 0 &&
		c.Y >= 0 &&
		c.X < tb.Ncols &&
		c.Y < tb.Nrows
}

func (self *LHTipbox) NRows() int {
	return self.Nrows
}

func (self *LHTipbox) NCols() int {
	return self.Ncols
}

func (tb *LHTipbox) GetChildByAddress(c WellCoords) LHObject {
	if !tb.AddressExists(c) {
		return nil
	}
	return tb.Tips[c.X][c.Y]
}

func (tb *LHTipbox) CoordsToWellCoords(idGen *id.IDGenerator, r Coordinates3D) (WellCoords, Coordinates3D) {
	//get relative Coordinates
	rel := r.Subtract(tb.GetPosition())
	tipSize := tb.Tiptype.GetSize()
	wc := WellCoords{
		int(math.Floor(((rel.X - tb.TipXStart + 0.5*tipSize.X) / tb.TipXOffset))),
		int(math.Floor(((rel.Y - tb.TipYStart + 0.5*tipSize.Y) / tb.TipYOffset))),
	}
	if wc.X < 0 {
		wc.X = 0
	} else if wc.X >= tb.Ncols {
		wc.X = tb.Ncols - 1
	}
	if wc.Y < 0 {
		wc.Y = 0
	} else if wc.Y >= tb.Nrows {
		wc.Y = tb.Nrows - 1
	}

	r2, _ := tb.WellCoordsToCoords(idGen, wc, TopReference)

	return wc, r.Subtract(r2)
}

func (tb *LHTipbox) WellCoordsToCoords(idGen *id.IDGenerator, wc WellCoords, r WellReference) (Coordinates3D, bool) {
	if !tb.AddressExists(wc) {
		return Coordinates3D{}, false
	}

	var z float64
	if r == BottomReference {
		z = tb.TipZStart
	} else if r == TopReference {
		z = tb.TipZStart + tb.Tiptype.GetSize().Z
	} else {
		return Coordinates3D{}, false
	}

	return tb.GetPosition().Add(Coordinates3D{
		tb.TipXStart + float64(wc.X)*tb.TipXOffset,
		tb.TipYStart + float64(wc.Y)*tb.TipYOffset,
		z}), true
}

//HasTipAt
func (tb *LHTipbox) HasTipAt(c WellCoords) bool {
	return tb.AddressExists(c) && tb.Tips[c.X][c.Y] != nil
}

//RemoveTip
func (tb *LHTipbox) RemoveTip(c WellCoords) *LHTip {
	if !tb.AddressExists(c) {
		return nil
	}
	tip := tb.Tips[c.X][c.Y]
	tb.Tips[c.X][c.Y] = nil
	return tip
}

//PutTip
func (tb *LHTipbox) PutTip(c WellCoords, tip *LHTip) bool {
	if !tb.AddressExists(c) {
		return false
	}
	if tb.HasTipAt(c) {
		return false
	}
	tb.Tips[c.X][c.Y] = tip
	return true
}

// Refill replace all the tips in the tipbox, leaving it full and clean again
func (tb *LHTipbox) Refill(idGen *id.IDGenerator) {
	initialize_tips(idGen, tb, tb.Tiptype)
}

func (tb *LHTipbox) MarshalJSON() ([]byte, error) {
	return json.Marshal(newSTipbox(tb))
}

func (tb *LHTipbox) UnmarshalJSON(data []byte) error {
	var stb sTipbox
	if err := json.Unmarshal(data, &stb); err != nil {
		return err
	}
	stb.Fill(tb)
	return nil
}

func initialize_tips(idGen *id.IDGenerator, tipbox *LHTipbox, tiptype *LHTip) *LHTipbox {
	nr := tipbox.Nrows
	nc := tipbox.Ncols
	//make sure tips are in the center of the address
	x_off := -tiptype.GetSize().X / 2.
	y_off := -tiptype.GetSize().Y / 2.
	for i := 0; i < nc; i++ {
		for j := 0; j < nr; j++ {
			tipbox.Tips[i][j] = tiptype.Dup(idGen)
			tipbox.Tips[i][j].SetOffset(Coordinates3D{ //nolint
				X: tipbox.TipXStart + float64(i)*tipbox.TipXOffset + x_off,
				Y: tipbox.TipYStart + float64(j)*tipbox.TipYOffset + y_off,
				Z: tipbox.TipZStart,
			})
			tipbox.Tips[i][j].SetParent(tipbox) //nolint
		}
	}
	tipbox.NTips = tipbox.Nrows * tipbox.Ncols
	return tipbox
}
