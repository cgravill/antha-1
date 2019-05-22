package tests

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/antha-lang/antha/antha/anthalib/mixer"
	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"github.com/antha-lang/antha/inventory/components"
	"github.com/antha-lang/antha/laboratory"
	"github.com/antha-lang/antha/laboratory/effects/id"
	"github.com/antha-lang/antha/laboratory/testlab"
	lh "github.com/antha-lang/antha/microArch/scheduler/liquidhandling"
	"github.com/pkg/errors"
)

type initFinalCmp struct {
	CNameI string
	CNameF string
	VolI   float64
	VolF   float64
}

func (ifc initFinalCmp) IsZero() bool {
	v := initFinalCmp{}
	return reflect.DeepEqual(v, ifc)
}

func getComponents(lab *laboratory.Laboratory, t *testing.T) (*wtype.Liquid, *wtype.Liquid, error) {
	cmp1, err := lab.Inventory.Components.NewComponent(components.WaterType)
	if err != nil {
		return nil, nil, err
	}
	cmp2, err := lab.Inventory.Components.NewComponent("dna_part")
	if err != nil {
		return nil, nil, err
	}
	return cmp1, cmp2, nil
}

/*
func TestBeforeVsAfterUserPlateMixInPlace(t *testing.T) {
	ctx := GetContextForTest()
	rq := makeRequest()
	lh := makeLiquidhandler(ctx)

	cmp1, cmp2 := getComponents(ctx, t)

	cmp1.Vol = 100.0
	cmp2.Vol = 50.0

	pl2, err := inventory.NewPlate(ctx, "pcrplate_skirted_riser20")

	if err != nil {
		t.Fatal(err)
	}

	pl2.Cols[0][0].AddComponent(cmp1)
	pl2.Cols[0][1].AddComponent(cmp2)

	mo := mixer.MixOptions{
		Inputs: []*wtype.Liquid{cmp1, cmp2},
	}

	ins := mixer.GenericMix(mo)

	rq.LHInstructions[ins.ID] = ins

	pl, err := inventory.NewPlate(ctx, "pcrplate_skirted_riser20")
	if err != nil {
		t.Fatal(err)
	}

	rq.InputPlatetypes = append(rq.InputPlatetypes, pl)

	rq.AddUserPlate(pl2)

	lh.Plan(ctx, rq)

	expected := make(map[string][]initFinalCmp)

	expected["dna_part"] = []initFinalCmp{initFinalCmp{CNameI: "dna_part", CNameF: "dna_part", VolI: 50.0, VolF: 5.0}}

	expected["water+dna_part"] = []initFinalCmp{initFinalCmp{CNameI: "water", CNameF: "water+dna_part", VolI: 100.0, VolF: 144.5}}

	compareInitFinalStates(t, lh, expected)

	fmt.Println("BEFORE")
	for _, p := range lh.Properties.Plates {
		fmt.Println(p.PlateName)
		p.OutputLayout()
	}
	fmt.Println("AFTER")
	for _, p := range lh.FinalProperties.Plates {
		fmt.Println(p.PlateName)
		p.OutputLayout()
	}
}
*/

func TestBeforeVsAfterUserPlateDest(t *testing.T) {
	testlab.WithTestLab(t, "", &testlab.TestElementCallbacks{
		Steps: func(lab *laboratory.Laboratory) error {

			rq := makeRequest(lab.IDGenerator)
			lh := makeLiquidhandler(lab)

			cmp1, cmp2, err := getComponents(lab, t)
			if err != nil {
				return err
			}

			cmp1.Vol = 100.0
			cmp2.Vol = 50.0

			pl2, err := lab.Inventory.Plates.NewPlate("pcrplate_skirted_riser20")

			if err != nil {
				return err
			}

			err = pl2.Cols[0][0].AddComponent(lab.IDGenerator, cmp1)
			if err != nil {
				return err
			}

			err = pl2.Cols[0][1].AddComponent(lab.IDGenerator, cmp2)
			if err != nil {
				return err
			}

			s1 := mixer.Sample(lab, cmp1, wunit.NewVolume(25.0, "ul"))
			s2 := mixer.Sample(lab, cmp2, wunit.NewVolume(10.0, "ul"))

			mo := mixer.MixOptions{
				Inputs:      []*wtype.Liquid{s1, s2},
				PlateType:   "pcrplate_skirted_riser20",
				Address:     "C1",
				Destination: pl2,
			}

			if err := rq.AddUserPlate(lab.IDGenerator, pl2); err != nil {
				return err
			}

			ins := mixer.GenericMix(lab, mo)

			rq.LHInstructions[ins.ID] = ins

			pl, err := lab.Inventory.Plates.NewPlate("pcrplate_skirted_riser20")
			if err != nil {
				return err
			}

			rq.InputPlatetypes = append(rq.InputPlatetypes, pl)

			if err := lh.Plan(lab.LaboratoryEffects, rq); err != nil {
				return err
			}

			expected := map[string][]initFinalCmp{
				"dna_part":                           {{CNameI: "dna_part", CNameF: "dna_part", VolI: 50.0, VolF: 39.5}},
				"0.286 v/v dna_part+0.714 v/v water": {{CNameI: "", CNameF: "0.286 v/v dna_part+0.714 v/v water", VolI: 0.0, VolF: 35.0}},
				"water":                              {{CNameI: "water", CNameF: "water", VolI: 100.0, VolF: 74.5}},
			}

			return compareInitFinalStates(lab, lh, expected)
		},
	})
}

func TestBeforeVsAfterUserPlateAutoDest(t *testing.T) {
	testlab.WithTestLab(t, "", &testlab.TestElementCallbacks{
		Steps: func(lab *laboratory.Laboratory) error {

			rq := makeRequest(lab.IDGenerator)
			lh := makeLiquidhandler(lab)

			cmp1, cmp2, err := getComponents(lab, t)
			if err != nil {
				return err
			}

			cmp1.Vol = 100.0
			cmp2.Vol = 50.0

			s1 := mixer.Sample(lab, cmp1, wunit.NewVolume(25.0, "ul"))
			s2 := mixer.Sample(lab, cmp2, wunit.NewVolume(10.0, "ul"))

			mo := mixer.MixOptions{
				Inputs: []*wtype.Liquid{s1, s2},
			}

			ins := mixer.GenericMix(lab, mo)

			rq.LHInstructions[ins.ID] = ins

			pl, err := lab.Inventory.Plates.NewPlate("pcrplate_skirted_riser20")
			if err != nil {
				return err
			}

			rq.InputPlatetypes = append(rq.InputPlatetypes, pl)

			pl2, err := lab.Inventory.Plates.NewPlate("pcrplate_skirted_riser20")
			if err != nil {
				return err
			}

			if err := pl2.Cols[0][0].AddComponent(lab.IDGenerator, cmp1); err != nil {
				return err
			}

			if err := pl2.Cols[0][1].AddComponent(lab.IDGenerator, cmp2); err != nil {
				return err
			}

			if err := rq.AddUserPlate(lab.IDGenerator, pl2); err != nil {
				return err
			}

			if err := lh.Plan(lab.LaboratoryEffects, rq); err != nil {
				return err
			}

			expected := map[string][]initFinalCmp{
				"dna_part":                           {{CNameI: "dna_part", CNameF: "dna_part", VolI: 50.0, VolF: 39.5}},
				"0.286 v/v dna_part+0.714 v/v water": {{CNameI: "", CNameF: "0.286 v/v dna_part+0.714 v/v water", VolI: 0.0, VolF: 35.0}},
				"water":                              {{CNameI: "water", CNameF: "water", VolI: 100.0, VolF: 74.5}},
			}

			return compareInitFinalStates(lab, lh, expected)
		},
	})
}

func TestBeforeVsAfterUserPlate(t *testing.T) {
	testlab.WithTestLab(t, "", &testlab.TestElementCallbacks{
		Steps: func(lab *laboratory.Laboratory) error {

			rq := makeRequest(lab.IDGenerator)
			lh := makeLiquidhandler(lab)

			cmp1, cmp2, err := getComponents(lab, t)
			if err != nil {
				return err
			}

			cmp1.Vol = 100.0
			cmp2.Vol = 50.0

			s1 := mixer.Sample(lab, cmp1, wunit.NewVolume(25.0, "ul"))
			s2 := mixer.Sample(lab, cmp2, wunit.NewVolume(10.0, "ul"))

			mo := mixer.MixOptions{
				Inputs:    []*wtype.Liquid{s1, s2},
				PlateType: "pcrplate_skirted_riser20",
				Address:   "C1",
				PlateNum:  1,
			}

			ins := mixer.GenericMix(lab, mo)

			rq.LHInstructions[ins.ID] = ins

			pl, err := lab.Inventory.Plates.NewPlate("pcrplate_skirted_riser20")
			if err != nil {
				return err
			}

			rq.InputPlatetypes = append(rq.InputPlatetypes, pl)

			pl2, err := lab.Inventory.Plates.NewPlate("pcrplate_skirted_riser20")
			if err != nil {
				return err
			}

			if err := pl2.Cols[0][0].AddComponent(lab.IDGenerator, cmp1); err != nil {
				return err
			}

			if err := pl2.Cols[0][1].AddComponent(lab.IDGenerator, cmp2); err != nil {
				return err
			}

			if err := rq.AddUserPlate(lab.IDGenerator, pl2); err != nil {
				return err
			}

			if err := lh.Plan(lab.LaboratoryEffects, rq); err != nil {
				return err
			}

			expected := map[string][]initFinalCmp{
				"dna_part":                           {{CNameI: "dna_part", CNameF: "dna_part", VolI: 50.0, VolF: 39.5}},
				"0.286 v/v dna_part+0.714 v/v water": {{CNameI: "", CNameF: "0.286 v/v dna_part+0.714 v/v water", VolI: 0.0, VolF: 35.0}},
				"water":                              {{CNameI: "water", CNameF: "water", VolI: 100.0, VolF: 74.5}},
			}

			return compareInitFinalStates(lab, lh, expected)
		},
	})
}

/*
func TestBeforeVsAfterMixInPlace(t *testing.T) {
	ctx := GetContextForTest()
	rq := makeRequest()
	lh := makeLiquidhandler(ctx)

	cmp1, cmp2 := getComponents(ctx, t)

	cmp1.Vol = 100.0
	cmp2.Vol = 50.0

	mo := mixer.MixOptions{
		Inputs: []*wtype.Liquid{cmp1, cmp2},
	}

	ins := mixer.GenericMix(mo)

	rq.LHInstructions[ins.ID] = ins

	pl, err := inventory.NewPlate(ctx, "pcrplate_skirted_riser20")
	if err != nil {
		t.Fatal(err)
	}

	rq.InputPlatetypes = append(rq.InputPlatetypes, pl)

	lh.Plan(ctx, rq)

	expected := make(map[string][]initFinalCmp)

	expected["dna_part"] = []initFinalCmp{initFinalCmp{CNameI: "dna_part", CNameF: "dna_part", VolI: 50.0, VolF: 5.0}}

	expected["water+dna_part"] = []initFinalCmp{initFinalCmp{CNameI: "water", CNameF: "water+dna_part", VolI: 100.0, VolF: 144.5}}

	compareInitFinalStates(t, lh, expected)
	fmt.Println("BEFORE")
	for _, p := range lh.Properties.Plates {
		fmt.Println(p.PlateName)
		p.OutputLayout()
	}
	fmt.Println("AFTER")
	for _, p := range lh.FinalProperties.Plates {
		fmt.Println(p.PlateName)
		p.OutputLayout()
	}
}
*/
func TestBeforeVsAfterAutoAllocateDest(t *testing.T) {
	testlab.WithTestLab(t, "", &testlab.TestElementCallbacks{
		Steps: func(lab *laboratory.Laboratory) error {

			rq := makeRequest(lab.IDGenerator)
			lh := makeLiquidhandler(lab)

			cmp1, cmp2, err := getComponents(lab, t)
			if err != nil {
				return err
			}

			s1 := mixer.Sample(lab, cmp1, wunit.NewVolume(50.0, "ul"))
			s2 := mixer.Sample(lab, cmp2, wunit.NewVolume(25.0, "ul"))

			mo := mixer.MixOptions{
				Inputs: []*wtype.Liquid{s1, s2},
			}

			ins := mixer.GenericMix(lab, mo)

			rq.LHInstructions[ins.ID] = ins

			pl, err := lab.Inventory.Plates.NewPlate("pcrplate_skirted_riser20")
			if err != nil {
				return err
			}

			rq.InputPlatetypes = append(rq.InputPlatetypes, pl)
			rq.OutputPlatetypes = append(rq.OutputPlatetypes, pl.Dup(lab.IDGenerator))

			if err := lh.Plan(lab.LaboratoryEffects, rq); err != nil {
				return err
			}

			expected := map[string][]initFinalCmp{
				"dna_part":                           {{CNameI: "dna_part", CNameF: "dna_part", VolI: 30.5, VolF: 5.0}},
				"water":                              {{CNameI: "water", CNameF: "water", VolI: 55.5, VolF: 5.0}},
				"0.333 v/v dna_part+0.667 v/v water": {{CNameI: "", CNameF: "0.333 v/v dna_part+0.667 v/v water", VolI: 0.0, VolF: 75.0}},
			}

			return compareInitFinalStates(lab, lh, expected)
		},
	})
}

func TestBeforeVsAfterAutoAllocate(t *testing.T) {
	testlab.WithTestLab(t, "", &testlab.TestElementCallbacks{
		Steps: func(lab *laboratory.Laboratory) error {

			rq := makeRequest(lab.IDGenerator)
			lh := makeLiquidhandler(lab)

			cmp1, cmp2, err := getComponents(lab, t)
			if err != nil {
				return err
			}

			s1 := mixer.Sample(lab, cmp1, wunit.NewVolume(50.0, "ul"))
			s2 := mixer.Sample(lab, cmp2, wunit.NewVolume(25.0, "ul"))

			mo := mixer.MixOptions{
				Inputs:    []*wtype.Liquid{s1, s2},
				PlateType: "pcrplate_skirted_riser20",
				Address:   "A1",
				PlateNum:  1,
			}

			ins := mixer.GenericMix(lab, mo)

			rq.LHInstructions[ins.ID] = ins

			pl, err := lab.Inventory.Plates.NewPlate("pcrplate_skirted_riser20")
			if err != nil {
				return err
			}

			rq.InputPlatetypes = append(rq.InputPlatetypes, pl)

			if err := lh.Plan(lab.LaboratoryEffects, rq); err != nil {
				return err
			}

			expected := map[string][]initFinalCmp{
				"dna_part":                           {{CNameI: "dna_part", CNameF: "dna_part", VolI: 30.5, VolF: 5.0}},
				"water":                              {{CNameI: "water", CNameF: "water", VolI: 55.5, VolF: 5.0}},
				"0.333 v/v dna_part+0.667 v/v water": {{CNameI: "", CNameF: "0.333 v/v dna_part+0.667 v/v water", VolI: 0.0, VolF: 75.0}},
			}

			return compareInitFinalStates(lab, lh, expected)
		},
	})
}

// all means just one or the whole lot
func del(a initFinalCmp, ar []initFinalCmp, all bool) []initFinalCmp {
	ar2 := make([]initFinalCmp, 0, len(ar)-1)
	d := false
	for _, b := range ar {
		if !reflect.DeepEqual(a, b) || (d && !all) {
			ar2 = append(ar2, b)
			d = true
		}
	}

	return ar2
}

/*

type initFinalCmp struct {
	CNameI string
	CNameF string
	VolI   float64
	VolF   float64
}

*/
func findWells(wi, wf *wtype.LHWell, ar []initFinalCmp) initFinalCmp {
	ifc := initFinalCmp{CNameI: wi.WContents.CName, CNameF: wf.WContents.CName, VolI: wi.WContents.Vol, VolF: wf.WContents.Vol}

	return findIFC(ifc, ar)
}

func findIFC(ifc initFinalCmp, ar []initFinalCmp) initFinalCmp {
	r := initFinalCmp{}

	for _, ifc2 := range ar {
		if reflect.DeepEqual(ifc, ifc2) {
			r = ifc2
			break
		}
	}

	return r
}

func compareInitFinalStates(lab *laboratory.Laboratory, lh *lh.Liquidhandler, expected map[string][]initFinalCmp) error {
	for _, pos := range lh.Properties.InputSearchPreferences() {
		p, ok := lh.Properties.Plates[pos]
		p2, ok2 := lh.FinalProperties.Plates[pos]

		if ok != ok2 {
			return errors.New("Plates moving in simple liquid handling plan")
		}

		if ok {
			// find each component and make sure it has stayed in the same place
			for _, crd := range p.AllWellPositions(false) {
				w := p.Wellcoords[crd]
				w2 := p2.Wellcoords[crd]

				e1 := w.IsEmpty(lab.IDGenerator)
				e2 := w2.IsEmpty(lab.IDGenerator)

				if e1 && e2 {
					continue
				}

				v, ok3 := expected[w2.WContents.CName]

				if ok3 {
					ifc := findWells(w, w2, v)

					if ifc.IsZero() {
						return fmt.Errorf("Extra components of type %s in before / after: \"%s\" %f \"%s\" %f", w2.WContents.CName, w.WContents.CName, w.WContents.Vol, w2.WContents.CName, w2.WContents.Vol)
					}

					// good, delete this now

					expected[w2.WContents.CName] = del(ifc, v, false)
				} else {
					return fmt.Errorf("Unexpected components in before / after: \"%s\" %f \"%s\" %f", w.WContents.CName, w.WContents.Vol, w2.WContents.CName, w2.WContents.Vol)
				}
			}
		}
	}

	// is anything left in the expected pile?

	for k, v := range expected {
		if len(v) != 0 {
			return fmt.Errorf("Unmatched components of type %s : %d total", k, len(v))
		}
	}
	return nil
}

func makeLiquidhandler(lab *laboratory.Laboratory) *lh.Liquidhandler {
	return lh.Init(makeGilson(lab))
}

func makeRequest(idGen *id.IDGenerator) *lh.LHRequest {
	return lh.NewLHRequest(idGen)
}
