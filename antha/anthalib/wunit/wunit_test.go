// wunit/wunit_test.go: Part of the Antha language
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
// contact license@antha-lang.org or write to the Antha team c/o
// Synthace Ltd. The London Bioscience Innovation Centre
// 2 Royal College St, London NW1 0NH UK

package wunit

import (
	"encoding/json"
	"fmt"
	"testing"
)

func StoreMeasurement(m Measurement) {
	// do something
}

func TestBasic(*testing.T) {
	ExampleBasic()
}
func TestTwo(*testing.T) {
	ExampleTwo()
}
func TestFour(*testing.T) {
	ExampleFour()
}
func TestFive(*testing.T) {
	ExampleFive()
}

func TestSIParsing(*testing.T) {
	ExampleSeven()
}

func TestUnitConversion(*testing.T) {
	ExampleEight()
}

func TestJSON(*testing.T) {
	ExampleNine()
}

func ExampleBasic() {
	degreeC := GenericPrefixedUnit{GenericUnit{"DegreeC", "C", 1.0, "C"}, SIPrefix{"m", 1e-03}}
	cm := ConcreteMeasurement{1.0, &degreeC}
	TdegreeC := Temperature{&cm}
	fmt.Println(TdegreeC.SIValue())
	// Output:
	// 0.001
}
func ExampleTwo() {
	Joule := GenericPrefixedUnit{GenericUnit{"Joule", "J", 1.0, "J"}, SIPrefix{"k", 1e3}}
	cm := ConcreteMeasurement{23.4, &Joule}
	NJoule := Energy{&cm}
	fmt.Println(NJoule.SIValue())
	// Output:
	// 23400
}

func ExampleFour() {
	p := NewPressure(56.2, "Pa")
	fmt.Println(p.RawValue())

	p.SetValue(34.0)

	fmt.Println(p.RawValue())

	// Output:
	// 56.2
	// 34
}

func ExampleFive() {

	fmt.Println(PrefixMul("m", "m"))

	// Output:
	// u
}

func ExampleSix() {
	fmt.Println(k)
	fmt.Println(G)
	fmt.Println(p)
	// Output:
	// 3
	// 9
	// -12

}

func ExampleSeven() {
	pu := ParsePrefixedUnit("GHz")
	fmt.Println(pu.Symbol())
	fmt.Println(pu.BaseSIConversionFactor())
	pu = ParsePrefixedUnit("uM")
	fmt.Println(pu.Symbol())
	fmt.Println(pu.BaseSIConversionFactor())
	// Output:
	// GHz
	// 1e+09
	// uM
	// 1e-06
}

func ExampleEight() {
	// testing the new conversion methods
	pu := ParsePrefixedUnit("GHz")
	pu2 := ParsePrefixedUnit("MHz")
	pu3 := ParsePrefixedUnit("l")
	meas := ConcreteMeasurement{10, pu}
	meas2 := ConcreteMeasurement{50, pu2}
	meas3 := ConcreteMeasurement{10, pu3}

	fmt.Println(meas.ToString(), " is ", meas.ConvertTo(meas.Unit()), " ", pu.PrefixedSymbol())
	fmt.Println(meas2.ToString(), " is ", meas2.ConvertTo(meas.Unit()), " ", pu.PrefixedSymbol())
	fmt.Println(meas2.ToString(), " is ", meas2.ConvertTo(meas2.Unit()), " ", pu2.PrefixedSymbol())
	fmt.Println(meas.ToString(), " is ", meas.ConvertTo(meas2.Unit()), " ", pu2.PrefixedSymbol())
	fmt.Println(meas3.ToString())
	fmt.Println(meas3.Unit().ToString())
	fmt.Println(pu3.PrefixedSymbol())
	// Output:
	// 10.000GHz  is  10   GHz
	// 50.000MHz  is  0.05   GHz
	// 50.000MHz  is  50   MHz
	// 10.000GHz  is  10000   MHz
	// 10.000l
	// Name: litre Symbol: l Conversion: 1    BaseUnit: l
	// l
}

func ExampleNine() {
	// testing JSON functions

	pu := ParsePrefixedUnit("GHz")
	//meas := ConcreteMeasurement{10, pu}

	x := PrefixedUnit(pu)

	b, err := json.Marshal(x)

	fmt.Println(string(b))
	fmt.Println(err)

	var pu2 GenericPrefixedUnit

	er2 := json.Unmarshal(b, &pu2)

	fmt.Println("Unmarshalled: ", pu2)
	fmt.Println(er2)

}

// simple reverse complement check to test testing methodology initially

type testunit struct {
	value        float64
	prefix       string
	unit         string
	prefixedunit string
	siresult     float64
}

var units = []testunit{
	{2.0000000000000003e-06, "", "l", "l", 2.0000000000000003e-06},
	{2.05, "u", "l", "ul", 2.05e-6},
}

var concs = []testunit{
	{2.0000000000000003e-06, "", "g/l", "g/l", 2.0000000000000003e-06},
	{2.0000000000000003e-06, "", "kg/l", "kg/l", 2.0000000000000005e-09},
	{2.05, "m", "g/l", "mg/l", 0.0020499999999999997},
	{2.05, "m", "Mol/l", "mMol/l", 0.0020499999999999997},
	{2.05, "m", "g/l", "ng/ul", 0.0020499999999999997},
}

type VolumeArithmetic struct {
	VolumeA    Volume
	VolumeB    Volume
	Sum        Volume
	Difference Volume
	Factor     float64
	Product    Volume
	Quotient   Volume
}

var volumearithmetictests = []VolumeArithmetic{
	{
		VolumeA:    NewVolume(1, "ul"),
		VolumeB:    NewVolume(1, "ul"),
		Sum:        NewVolume(2, "ul"),
		Difference: NewVolume(0, "ul"),
		Factor:     1.0,
		Product:    NewVolume(1, "ul"),
		Quotient:   NewVolume(1, "ul"),
	},
	{
		VolumeA:    NewVolume(100, "ul"),
		VolumeB:    NewVolume(10, "ul"),
		Sum:        NewVolume(110, "ul"),
		Difference: NewVolume(90, "ul"),
		Factor:     10.0,
		Product:    NewVolume(1000, "ul"),
		Quotient:   NewVolume(10, "ul"),
	},
	{
		VolumeA:    NewVolume(1000000, "ul"),
		VolumeB:    NewVolume(10, "ul"),
		Sum:        NewVolume(1000010, "ul"),
		Difference: NewVolume(999990, "ul"),
		Factor:     10.0,
		Product:    NewVolume(10000000, "ul"),
		Quotient:   NewVolume(100000, "ul"),
	},
	{
		VolumeA:    NewVolume(1000, "ml"),
		VolumeB:    NewVolume(10, "ul"),
		Sum:        NewVolume(1000010, "ul"),
		Difference: NewVolume(999990, "ul"),
		Factor:     10.0,
		Product:    NewVolume(10000000, "ul"),
		Quotient:   NewVolume(100000, "ul"),
	},
}

func TestSubstractVolumes(t *testing.T) {
	for _, testunit := range volumearithmetictests {
		r := SubtractVolumes(testunit.VolumeA, []Volume{testunit.VolumeB})
		if r.SIValue() != testunit.Difference.SIValue() {
			t.Error(
				"For", testunit.VolumeA, "/n",
				"expected", testunit.Difference, "\n",
				"got", r, "\n",
			)
		}
	}

}

func TestAddVolumes(t *testing.T) {
	for _, testunit := range volumearithmetictests {
		r := AddVolumes([]Volume{testunit.VolumeA, testunit.VolumeB})
		if r.SIValue() != testunit.Sum.SIValue() {
			t.Error(
				"For", testunit.VolumeA, "/n",
				"expected", testunit.Sum, "\n",
				"got", r, "\n",
			)
		}
	}

}

func TestMultiplyVolumes(t *testing.T) {
	for _, testunit := range volumearithmetictests {
		r := MultiplyVolume(testunit.VolumeA, testunit.Factor)
		if r.SIValue() != testunit.Product.SIValue() {
			t.Error(
				"For", testunit.VolumeA, "/n",
				"expected", testunit.Product, "\n",
				"got", r, "\n",
			)
		}
	}

}

func TestDivideVolumes(t *testing.T) {
	for _, testunit := range volumearithmetictests {
		r := DivideVolume(testunit.VolumeA, testunit.Factor)
		if r.SIValue() != testunit.Quotient.SIValue() {
			t.Error(
				"For", testunit.VolumeA, "/n",
				"expected", testunit.Quotient, "\n",
				"got", r, "\n",
			)
		}
	}

}

type ConcArithmetic struct {
	ValueA     Concentration
	ValueB     Concentration
	Sum        Concentration
	Difference Concentration
	Factor     float64
	Product    Concentration
	Quotient   Concentration
}

var concarithmetictests = []ConcArithmetic{
	{
		ValueA:     NewConcentration(1, "ng/ul"),
		ValueB:     NewConcentration(1, "ng/ul"),
		Sum:        NewConcentration(2, "ng/ul"),
		Difference: NewConcentration(0, "ng/ul"),
		Factor:     1.0,
		Product:    NewConcentration(1, "ng/ul"),
		Quotient:   NewConcentration(1, "ng/ul"),
	},
	{
		ValueA:     NewConcentration(100, "ng/ul"),
		ValueB:     NewConcentration(10, "ng/ul"),
		Sum:        NewConcentration(110, "ng/ul"),
		Difference: NewConcentration(90, "ng/ul"),
		Factor:     10.0,
		Product:    NewConcentration(1000, "ng/ul"),
		Quotient:   NewConcentration(10, "ng/ul"),
	},
	{
		ValueA:     NewConcentration(1000000, "mg/l"),
		ValueB:     NewConcentration(10, "ng/ul"),
		Sum:        NewConcentration(1000010, "ng/ul"),
		Difference: NewConcentration(999990, "ng/ul"),
		Factor:     10.0,
		Product:    NewConcentration(10000000, "ng/ul"),
		Quotient:   NewConcentration(100000, "ng/ul"),
	},
	{
		ValueA:     NewConcentration(1000, "g/l"),
		ValueB:     NewConcentration(10, "ng/ul"),
		Sum:        NewConcentration(1000010, "ng/ul"),
		Difference: NewConcentration(999990, "ng/ul"),
		Factor:     10.0,
		Product:    NewConcentration(10000000, "ng/ul"),
		Quotient:   NewConcentration(100000, "ng/ul"),
	},
}

func TestMultiplyConcentrations(t *testing.T) {
	for _, testunit := range concarithmetictests {
		r := MultiplyConcentration(testunit.ValueA, testunit.Factor)
		if r.SIValue() != testunit.Product.SIValue() {
			t.Error(
				"For", testunit.ValueA, "/n",
				"expected", testunit.Product, "\n",
				"got", r, "\n",
			)
		}
	}

}

func TestDivideConcentration(t *testing.T) {
	for _, testunit := range concarithmetictests {
		r := DivideConcentration(testunit.ValueA, testunit.Factor)
		if r.SIValue() != testunit.Quotient.SIValue() {
			t.Error(
				"For", testunit.ValueA, "/n",
				"expected", testunit.Quotient, "\n",
				"got", r, "\n",
			)
		}
	}

}

func TestNewMeasurement(t *testing.T) {
	for _, testunit := range units {
		r := NewMeasurement(testunit.value, testunit.prefix, testunit.unit)
		if r.SIValue() != testunit.siresult {
			t.Error(
				"For", testunit.value, testunit.prefix, testunit.unit, "/n",
				"expected", testunit.siresult, "\n",
				"got", r.SIValue(), "\n",
			)
		}
	}

}

func TestNewVolume(t *testing.T) {
	for _, testunit := range units {
		r := NewVolume(testunit.value, testunit.prefixedunit)
		if r.SIValue() != testunit.siresult {
			t.Error(
				"For", testunit.value, testunit.prefixedunit, "/n",
				"expected", testunit.siresult, "\n",
				"got", r.SIValue(), "\n",
			)
		}
	}

}

func TestNewConcentration(t *testing.T) {
	for _, testunit := range concs {
		r := NewConcentration(testunit.value, testunit.prefixedunit)
		if r.SIValue() != testunit.siresult {
			t.Error(
				"For", testunit.value, testunit.prefixedunit, "/n",
				"expected", testunit.siresult, "\n",
				"got", r.SIValue(), "\n",
			)
		}
	}

}
