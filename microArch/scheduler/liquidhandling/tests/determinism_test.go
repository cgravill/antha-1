package tests

import (
	"github.com/Synthace/antha/antha/anthalib/wtype"
	"github.com/Synthace/antha/antha/anthalib/wunit"
	"github.com/Synthace/antha/laboratory"
)

func GetComponentForTest(lab *laboratory.Laboratory, name string, vol wunit.Volume) *wtype.Liquid {
	c, err := lab.Inventory.Components.NewComponent(name)
	if err != nil {
		panic(err)
	}
	c.SetVolume(vol)
	return c
}
