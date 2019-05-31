package tests

import (
	"fmt"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"github.com/antha-lang/antha/inventory/components"
	"github.com/antha-lang/antha/laboratory"
	"github.com/antha-lang/antha/microArch/driver/liquidhandling"
	"github.com/antha-lang/antha/workflow"
)

const (
	HVMinRate = 0.225
	HVMaxRate = 37.5
	LVMinRate = 0.0225
	LVMaxRate = 3.75
)

func MakeGilsonWithPlatesAndTipboxesForTest(lab *laboratory.Laboratory, inputPlateType wtype.PlateTypeName) *liquidhandling.LHProperties { //nolint
	ret, err := makeGilsonWithPlatesAndTipboxesForTest(lab, inputPlateType)
	if err != nil {
		panic(err)
	}
	return ret
}

func MakeGilsonWithTipboxesForTest(lab *laboratory.Laboratory) *liquidhandling.LHProperties { //nolint
	ret, err := makeGilsonWithTipboxesForTest(lab)
	if err != nil {
		panic(err)
	}
	return ret
}

func getHVConfig() *wtype.LHChannelParameter {
	minvol := wunit.NewVolume(20, "ul")
	maxvol := wunit.NewVolume(200, "ul")
	minspd := wunit.NewFlowRate(HVMinRate, "ml/min")
	maxspd := wunit.NewFlowRate(HVMaxRate, "ml/min")

	return wtype.NewLHChannelParameter("HVconfig", "GilsonPipetmax", minvol, maxvol, minspd, maxspd, 8, false, wtype.LHVChannel, 0)
}

func getLVConfig() *wtype.LHChannelParameter {
	newminvol := wunit.NewVolume(0.5, "ul")
	newmaxvol := wunit.NewVolume(20, "ul")
	newminspd := wunit.NewFlowRate(LVMinRate, "ml/min")
	newmaxspd := wunit.NewFlowRate(LVMaxRate, "ml/min")

	return wtype.NewLHChannelParameter("LVconfig", "GilsonPipetmax", newminvol, newmaxvol, newminspd, newmaxspd, 8, false, wtype.LHVChannel, 1)
}

func MakeGilsonForTest(lab *laboratory.Laboratory, tipList []wtype.TipType) *liquidhandling.LHProperties {
	// gilson pipetmax

	layout := make(map[string]*wtype.LHPosition)
	i := 0
	x0 := 3.886
	y0 := 3.513
	z0 := -82.035
	xi := 149.86
	yi := 95.25
	xp := x0 // nolint
	yp := y0
	zp := z0
	for y := 0; y < 3; y++ {
		xp = x0
		for x := 0; x < 3; x++ {
			pos := wtype.NewLHPosition(fmt.Sprintf("position_%d", i+1), wtype.Coordinates3D{X: xp, Y: yp, Z: zp}, wtype.SBSFootprint)
			layout[pos.Name] = pos
			i += 1
			xp += xi
		}
		yp += yi
	}
	lhp := liquidhandling.NewLHProperties(lab.IDGenerator, "Pipetmax", "Gilson",
		liquidhandling.LLLiquidHandler, liquidhandling.DisposableTips, layout)
	// get tips permissible from the factory
	SetUpTipsFor(lab, lhp, tipList)

	lhp.Preferences = &workflow.LayoutOpt{
		Tipboxes:  workflow.Addresses{"position_2", "position_3", "position_6", "position_9", "position_8", "position_5", "position_4", "position_7"},
		Inputs:    workflow.Addresses{"position_4", "position_5", "position_6", "position_9", "position_8", "position_3"},
		Outputs:   workflow.Addresses{"position_8", "position_9", "position_6", "position_5", "position_3", "position_1"},
		Washes:    workflow.Addresses{"position_8"},
		Tipwastes: workflow.Addresses{"position_1", "position_7"},
		Wastes:    workflow.Addresses{"position_9"},
	}

	hvconfig := getHVConfig()
	hvadaptor := wtype.NewLHAdaptor(lab.IDGenerator, "DummyAdaptor", "Gilson", hvconfig)
	hvhead := wtype.NewLHHead(lab.IDGenerator, "HVHead", "Gilson", hvconfig)
	hvhead.Adaptor = hvadaptor

	lvconfig := getLVConfig()
	lvadaptor := wtype.NewLHAdaptor(lab.IDGenerator, "DummyAdaptor", "Gilson", lvconfig)
	lvhead := wtype.NewLHHead(lab.IDGenerator, "LVHead", "Gilson", lvconfig)
	lvhead.Adaptor = lvadaptor

	ha := wtype.NewLHHeadAssembly(nil)
	ha.AddPosition(wtype.Coordinates3D{X: 0, Y: -18.08, Z: 0})
	ha.AddPosition(wtype.Coordinates3D{X: 0, Y: 0, Z: 0})
	if err := ha.LoadHead(hvhead); err != nil {
		panic(err)
	}
	if err := ha.LoadHead(lvhead); err != nil {
		panic(err)
	}
	lhp.Heads = append(lhp.Heads, hvhead, lvhead)
	lhp.Adaptors = append(lhp.Adaptors, hvadaptor, lvadaptor)
	lhp.HeadAssemblies = append(lhp.HeadAssemblies, ha)

	return lhp
}

func SetUpTipsFor(lab *laboratory.Laboratory, lhp *liquidhandling.LHProperties, tipList []wtype.TipType) *liquidhandling.LHProperties {
	tipTypeEnabled := make(map[wtype.TipType]bool, len(tipList))
	for _, tt := range tipList {
		tipTypeEnabled[tt] = true
	}

	seen := make(map[string]bool)

	err := lab.Inventory.TipBoxes.ForEach(func(tb wtype.LHTipbox) error {
		if tb.Mnfr == lhp.Mnfr || lhp.Mnfr == "MotherNature" {
			//ignore filter tips and the hacky "low volume high volume" ones
			//		if tb.Tiptype.Filtered || tb.Tiptype.Type == "LVGilson200" {
			//			continue
			//		}

			// ignore tips not in the list

			if tipTypeEnabled[tb.Tiptype.Type] {
				tip := tb.Tips[0][0]
				str := tip.Mnfr + string(tip.Type) + tip.MinVol.ToString() + tip.MaxVol.ToString()
				if seen[str] {
					return nil
				}

				seen[str] = true
				lhp.Tips = append(lhp.Tips, tb.Tips[0][0])
			}
		}
		return nil
	})
	if err != nil {
		panic(err) // it really can't be != nil because the above fun never returns a non nil err
	}
	return lhp
}

func defaultTipList() []wtype.TipType {
	return []wtype.TipType{"Gilson20", "Gilson200"}
}

func makeGilsonWithTipboxesForTest(lab *laboratory.Laboratory) (*liquidhandling.LHProperties, error) {
	params := MakeGilsonForTest(lab, defaultTipList())

	if tw, err := lab.Inventory.TipWastes.NewTipwaste("Gilsontipwaste"); err != nil {
		return nil, err
	} else if err := params.AddTipWaste(tw); err != nil {
		return nil, err
	}

	if tb, err := lab.Inventory.TipBoxes.NewTipbox("DL10 Tip Rack (PIPETMAX 8x20)"); err != nil {
		return nil, err
	} else if err := params.AddTipBox(tb); err != nil {
		return nil, err
	}

	if tb, err := lab.Inventory.TipBoxes.NewTipbox("DF200 Tip Rack (PIPETMAX 8x200)"); err != nil {
		return nil, err
	} else if err := params.AddTipBox(tb); err != nil {
		return nil, err
	}

	return params, nil
}

func makeGilsonWithPlatesAndTipboxesForTest(lab *laboratory.Laboratory, inputPlateType wtype.PlateTypeName) (*liquidhandling.LHProperties, error) {
	params, err := makeGilsonWithTipboxesForTest(lab)
	if err != nil {
		return nil, err
	}

	inputPlate, err := makeTestInputPlate(lab, inputPlateType)

	if err != nil {
		return nil, err
	}

	err = params.AddInputPlate(inputPlate)

	if err != nil {
		return nil, err
	}

	outputPlate, err := makeTestOutputPlate(lab)

	if err != nil {
		return nil, err
	}

	err = params.AddOutputPlate(outputPlate)

	if err != nil {
		return nil, err
	}
	return params, nil
}

func makeTestInputPlate(lab *laboratory.Laboratory, inputPlateType wtype.PlateTypeName) (*wtype.Plate, error) {
	if inputPlateType == "" {
		inputPlateType = "DWST12"
	}

	p, err := lab.Inventory.Plates.NewPlate(inputPlateType)

	if err != nil {
		return nil, err
	}

	c, err := lab.Inventory.Components.NewComponent(components.WaterType)

	if err != nil {
		return nil, err
	}

	c.Vol = 5000.0 // ul

	if _, err := p.AddComponent(lab.IDGenerator, c, true); err != nil {
		return nil, err
	}

	return p, nil
}

func makeTestOutputPlate(lab *laboratory.Laboratory) (*wtype.Plate, error) {
	return lab.Inventory.Plates.NewPlate("DSW96")
}
