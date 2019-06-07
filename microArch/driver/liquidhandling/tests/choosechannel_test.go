package tests

import (
	"errors"
	"fmt"
	"testing"

	"github.com/Synthace/antha/antha/anthalib/wtype"
	"github.com/Synthace/antha/antha/anthalib/wunit"
	"github.com/Synthace/antha/laboratory"
	"github.com/Synthace/antha/laboratory/testlab"
	"github.com/Synthace/antha/microArch/driver/liquidhandling"
)

func getVols() []wunit.Volume {
	// a selection of volumes
	vols := make([]wunit.Volume, 0, 1)
	for _, v := range []float64{0.5, 1.0, 2.0, 5.0, 10.0, 20.0, 30.0, 50.0, 100.0, 200.0} {
		vol := wunit.NewVolume(v, "ul")
		vols = append(vols, vol)
	}
	return vols
}

// answers to test

func getMinvols1() []wunit.Volume {
	v1 := wunit.NewVolume(0.5, "ul")
	v2 := wunit.NewVolume(20.0, "ul")

	ret := []wunit.Volume{v1, v1, v1, v1, v1, v1, v2, v2, v2, v2}

	return ret
}

func getMaxvols1() []wunit.Volume {
	v1 := wunit.NewVolume(20.0, "ul")
	v2 := wunit.NewVolume(200.0, "ul")

	ret := []wunit.Volume{v1, v1, v1, v1, v1, v1, v2, v2, v2, v2}

	return ret
}

func getVols2() []wunit.Volume {
	// a selection of volumes
	vols := make([]wunit.Volume, 0, 1)
	for _, v := range []float64{1.0, 2.0, 5.0, 10.0, 20.0, 30.0, 50.0, 100.0, 200.0} {
		vol := wunit.NewVolume(v, "ul")
		vols = append(vols, vol)
	}
	return vols
}

// answers to test

func getMinvols2() []wunit.Volume {
	v1 := wunit.NewVolume(0.5, "ul")

	ret := []wunit.Volume{v1, v1, v1, v1, v1, v1, v1, v1, v1}

	return ret
}

func getMaxvols2() []wunit.Volume {
	v1 := wunit.NewVolume(20.0, "ul")

	ret := []wunit.Volume{v1, v1, v1, v1, v1, v1, v1, v1, v1}

	return ret
}

func TestDefaultChooser(t *testing.T) {
	testlab.WithTestLab(t, "", &testlab.TestElementCallbacks{
		Steps: func(lab *laboratory.Laboratory) error {
			vols := getVols()
			lhp := MakeGilsonForTest(lab, defaultTipList())
			minvols := getMinvols1()
			maxvols := getMaxvols1()
			types := []wtype.TipType{"Gilson20", "Gilson20", "Gilson20", "Gilson20", "Gilson20", "Gilson20", "Gilson200", "Gilson200", "Gilson200", "Gilson200"}

			for i, vol := range vols {
				prm, tip, err := liquidhandling.ChooseChannel(vol, lhp)
				if err != nil {
					return err
				}

				var tiptype wtype.TipType

				if tip != nil {
					tiptype = tip.Type
				}

				mxr := maxvols[i]
				mnr := minvols[i]
				tpr := types[i]

				if prm == nil {
					if !mxr.IsZero() || !mnr.IsZero() || tpr != tiptype {
						return fmt.Errorf("Incorrect channel choice for volume %v\n\tGot nil want: %v %v ", vol.ToString(), mnr.ToString(), tpr)
					}

				} else if !prm.Maxvol.EqualTo(mxr) || !prm.Minvol.EqualTo(mnr) || tiptype != tpr {
					return fmt.Errorf("Incorrect channel choice for volume %v\n\tGot %v %v %v\n\tWant %v %v %v",
						vol.ToString(), prm.Minvol.ToString(), prm.Maxvol.ToString(), tiptype, mnr.ToString(), mxr.ToString(), tpr)
				}
			}
			return nil
		},
	})
}

func TestHVHVHVLVChooser(t *testing.T) {
	testlab.WithTestLab(t, "", &testlab.TestElementCallbacks{
		Steps: func(lab *laboratory.Laboratory) error {
			vols := getVols2()
			lhp := MakeGilsonForTest(lab, []wtype.TipType{"LVGilson200"})
			minvols := getMinvols2()
			maxvols := getMaxvols2()
			types := []wtype.TipType{"LVGilson200", "LVGilson200", "LVGilson200", "LVGilson200", "LVGilson200", "LVGilson200", "LVGilson200", "LVGilson200", "LVGilson200"}

			for i, vol := range vols {
				prm, tip, err := liquidhandling.ChooseChannel(vol, lhp)
				if err != nil {
					return err
				}

				var tiptype wtype.TipType

				if tip != nil {
					tiptype = tip.Type
				}

				mxr := maxvols[i]
				mnr := minvols[i]
				tpr := types[i]

				if prm == nil {
					if !mxr.IsZero() || !mnr.IsZero() || tpr != tiptype {
						return fmt.Errorf("Incorrect channel choice for volume %v\n\tGot nit want: %v %v", vol.ToString(), mnr.ToString(), tpr)
					}

				} else if !prm.Maxvol.EqualTo(mxr) || !prm.Minvol.EqualTo(mnr) || tiptype != tpr {
					return fmt.Errorf("Incorrect channel choice for volume %v\n\tGot %v %v %v\n\tWant %v %v %v",
						vol.ToString(), prm.Minvol.ToString(), prm.Maxvol.ToString(), tiptype, mnr.ToString(), mxr.ToString(), tpr)
				}
			}
			return nil
		},
	})
}

func TestSmallVolumeError(t *testing.T) {
	testlab.WithTestLab(t, "", &testlab.TestElementCallbacks{
		Steps: func(lab *laboratory.Laboratory) error {
			lhp := MakeGilsonForTest(lab, defaultTipList())

			vol := wunit.NewVolume(0.47, "ul")

			prm, tip, err := liquidhandling.ChooseChannel(vol, lhp)

			if prm != nil {
				return errors.New("channel was not nil for small volume")
			}
			if tip != nil {
				return errors.New("tip was not nil for small volume")
			}
			if err == nil {
				return errors.New("error not generated for small volume")
			}
			return nil
		},
	})
}
