// /anthalib/driver/incubating/driver.go: Part of the Antha language
// Copyright (C) 2015 The Antha authors. All rights reserved.
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

package incubating

import (
	"github.com/Synthace/antha/antha/anthalib/wtype"
	"github.com/Synthace/antha/antha/anthalib/wunit"
	"github.com/Synthace/antha/microArch/driver"
)

type IncubatingDriver interface {
	Initialize() driver.CommandStatus
	Finalize() driver.CommandStatus
	Incubate(what *wtype.LHSolution, temp wunit.Temperature, time wunit.Time, shaking bool) driver.CommandStatus
}

// change returns to driver.CommandStatus
//ShakeIncubate(what *wtype.LHSolution, temp wunit.Temperature, time wunit.Time, rpm int)
//Incubate(what *wtype.LHSolution, temp wunit.Temperature, time wunit.Time, shaking bool)
type ShakerIncubatordriver interface {
	Initialise()
	Command(string) string //driver.CommandStatus
	Shake(speed int) driver.CommandStatus
	ShakeforEngParameter(parameter string, liquid string, target float64)
	ShakeforTime(speed, time int) (status driver.CommandStatus)
	GetRemainingTime() driver.CommandStatus
	ShakeOff() driver.CommandStatus
	TempState() (onoroff string, settemp string, tempactual string)
	Temp(float64) driver.CommandStatus
	TempOff() driver.CommandStatus
	Open() driver.CommandStatus
	Close() driver.CommandStatus
	HomePos()
	LookupProperty(string) float64 // map lookup function from devices map in anthastandardlibrary. E.g. Maxspeed, dimensions etc..
	AllProperties() (map[string]float64, string)
	CheckState()
	//ShakeIncubate(what *wtype.LHSolution, temp wunit.Temperature, time wunit.Time, rpm int)
	//Incubate(what *wtype.LHSolution, temp wunit.Temperature, time wunit.Time, shaking bool)
}
