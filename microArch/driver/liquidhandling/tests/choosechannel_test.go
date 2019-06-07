package tests

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
	"testing"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"github.com/antha-lang/antha/laboratory/effects/id"
	"github.com/antha-lang/antha/microArch/driver/liquidhandling"
)

func NewChooseChannelsSingleTestTip(idGen *id.IDGenerator, name string, minUl, maxUl float64) *wtype.LHTip {
	shp := wtype.NewShape(wtype.CylinderShape, "mm", 1.0, 1.0, 1.0)
	return wtype.NewLHTip(idGen, "ACME test tips", wtype.TipType(name), minUl, maxUl, "ul", false, shp, 0.0)
}

func NewChooseChannelsSingleTestHead(idGen *id.IDGenerator, head int, minUl, maxUl float64) *wtype.LHHeadAssembly {
	cp := wtype.NewLHChannelParameter("testhead", "testplatform", wunit.NewVolume(minUl, "ul"), wunit.NewVolume(maxUl, "ul"), wunit.NewFlowRate(0.0, "ul/s"), wunit.NewFlowRate(0.0, "ul/s"), 8, false, wtype.LHHChannel, head)
	return &wtype.LHHeadAssembly{
		Positions: []*wtype.LHHeadAssemblyPosition{
			{
				Head: wtype.NewLHHead(idGen, "testhead", "ACME test heads", cp),
			},
		},
	}
}

type ChooseChannelsTest struct {
	Volumes       []wunit.Volume
	ExpectedHeads []int
	ExpectedTips  []string
	ShouldError   bool
}

func NewChooseChannelsSingleTest(volUl float64, head int, tip string, shouldErr bool) *ChooseChannelsTest {
	return &ChooseChannelsTest{
		Volumes:       []wunit.Volume{wunit.NewVolume(volUl, "ul")},
		ExpectedHeads: []int{head},
		ExpectedTips:  []string{tip},
		ShouldError:   shouldErr,
	}
}

func NewChooseChannelsTest(volsUl []float64, heads []int, tips []string, shouldErr bool) *ChooseChannelsTest {
	vols := make([]wunit.Volume, len(volsUl))
	for i, v := range volsUl {
		vols[i] = wunit.NewVolume(v, "ul")
	}

	return &ChooseChannelsTest{
		Volumes:       vols,
		ExpectedHeads: heads,
		ExpectedTips:  tips,
		ShouldError:   shouldErr,
	}
}

func (cct *ChooseChannelsTest) Run(params *liquidhandling.LHProperties) error {
	if prms, tips, err := liquidhandling.ChooseChannels(cct.Volumes, params); cct.ShouldError != (err != nil) {
		return errors.Errorf("errors didn't match: should error %t, got error: %s", cct.ShouldError, err)
	} else if err == nil {
		if len(prms) != len(tips) || len(tips) != len(cct.Volumes) {
			return errors.Errorf("ChooseChannels returned the wrong lengths: %d != %d != %d", len(cct.Volumes), len(prms), len(tips))
		}
		got := make([]string, len(cct.Volumes))
		expected := make([]string, len(cct.Volumes))
		for i := range cct.Volumes {
			if prms[i] != nil {
				got[i] = fmt.Sprintf("%s on %d", tips[i], prms[i].Head)
			}
			if cct.ExpectedTips[i] != "" {
				expected[i] = fmt.Sprintf("%s on %d", cct.ExpectedTips[i], cct.ExpectedHeads[i])
			}
		}

		if !reflect.DeepEqual(expected, got) {
			return errors.Errorf("tip or channel choice didn't match:\ne: %q\n g: %q", expected, got)
		}
	}
	return nil
}

type ChooseChannelsTests struct {
	Tips  []*wtype.LHTip
	Heads []*wtype.LHHeadAssembly
	Tests []*ChooseChannelsTest
}

func (cct *ChooseChannelsTests) Run(t *testing.T) {
	params := &liquidhandling.LHProperties{
		Tips:           cct.Tips,
		HeadAssemblies: cct.Heads,
	}

	for _, test := range cct.Tests {
		t.Run(fmt.Sprintf("%v", test.Volumes), func(t *testing.T) {
			if err := test.Run(params); err != nil {
				t.Error(err)
			}
		})
	}

	// also run the tests in reverse order to test order dependence
	tips := make([]*wtype.LHTip, len(cct.Tips))
	for i, tip := range cct.Tips {
		tips[len(tips)-1-i] = tip
	}
	heads := make([]*wtype.LHHeadAssembly, len(cct.Heads))
	for i, head := range cct.Heads {
		heads[len(heads)-1-i] = head
	}
	params = &liquidhandling.LHProperties{
		Tips:           tips,
		HeadAssemblies: heads,
	}

	for _, test := range cct.Tests {
		t.Run(fmt.Sprintf("%v_reversed", test.Volumes), func(t *testing.T) {
			if err := test.Run(params); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestChannelChoiceGilson(t *testing.T) {
	idGen := id.NewIDGenerator("ChooseChannelGilson")
	(&ChooseChannelsTests{
		Tips: []*wtype.LHTip{
			NewChooseChannelsSingleTestTip(idGen, "Low", 0.5, 20.0),
			NewChooseChannelsSingleTestTip(idGen, "High", 20.0, 200.0),
		},
		Heads: []*wtype.LHHeadAssembly{
			NewChooseChannelsSingleTestHead(idGen, 0, 0.5, 20.0),
			NewChooseChannelsSingleTestHead(idGen, 1, 20.0, 250.0),
		},
		Tests: []*ChooseChannelsTest{
			NewChooseChannelsSingleTest(0.47, 0, "", true),
			NewChooseChannelsTest([]float64{0.5, 0.5, 0.5, 0.5}, []int{0, 0, 0, 0}, []string{"Low", "Low", "Low", "Low"}, false),
			NewChooseChannelsTest([]float64{1.0, 1.0, 1.0, 1.0}, []int{0, 0, 0, 0}, []string{"Low", "Low", "Low", "Low"}, false),
			NewChooseChannelsTest([]float64{2.0, 2.0, 2.0, 2.0}, []int{0, 0, 0, 0}, []string{"Low", "Low", "Low", "Low"}, false),
			NewChooseChannelsTest([]float64{5.0, 5.0, 5.0, 5.0}, []int{0, 0, 0, 0}, []string{"Low", "Low", "Low", "Low"}, false),
			NewChooseChannelsTest([]float64{10.0, 10.0, 10.0, 10.0}, []int{0, 0, 0, 0}, []string{"Low", "Low", "Low", "Low"}, false),
			NewChooseChannelsTest([]float64{20.0, 20.0, 20.0, 20.0}, []int{0, 0, 0, 0}, []string{"Low", "Low", "Low", "Low"}, false),
			NewChooseChannelsTest([]float64{30.0, 30.0, 30.0, 30.0}, []int{1, 1, 1, 1}, []string{"High", "High", "High", "High"}, false),
			NewChooseChannelsTest([]float64{50.0, 50.0, 50.0, 50.0}, []int{1, 1, 1, 1}, []string{"High", "High", "High", "High"}, false),
			NewChooseChannelsTest([]float64{100.0, 100.0, 100.0, 100.0}, []int{1, 1, 1, 1}, []string{"High", "High", "High", "High"}, false),
			NewChooseChannelsTest([]float64{200.0, 200.0, 200.0, 200.0}, []int{1, 1, 1, 1}, []string{"High", "High", "High", "High"}, false),
			NewChooseChannelsTest([]float64{300.0, 300.0, 300.0, 300.0}, []int{1, 1, 1, 1}, []string{"High", "High", "High", "High"}, false),
			NewChooseChannelsTest([]float64{0.5, 0.5, 0.5, 0.5}, []int{0, 0, 0, 0}, []string{"Low", "Low", "Low", "Low"}, false),
			// inconsistent volumes will end up choosing the smaller tip since the larger tip cannot move 10ul
			NewChooseChannelsTest([]float64{10.0, 20.0, 30.0, 40.0}, []int{0, 0, 0, 0}, []string{"Low", "Low", "Low", "Low"}, false),
		},
	}).Run(t)
}

// TestChannelChoiceOverlapLow overlapping tips with the same minimum
func TestChannelChoiceOverlapLow(t *testing.T) {
	idGen := id.NewIDGenerator("ChooseChannelHamilton")
	(&ChooseChannelsTests{
		Tips: []*wtype.LHTip{
			NewChooseChannelsSingleTestTip(idGen, "Low", 0.5, 50.0),
			NewChooseChannelsSingleTestTip(idGen, "Standard", 0.5, 300.0),
			NewChooseChannelsSingleTestTip(idGen, "High", 20.0, 1000.0),
		},
		Heads: []*wtype.LHHeadAssembly{
			NewChooseChannelsSingleTestHead(idGen, 0, 0.5, 1000.0),
		},
		Tests: []*ChooseChannelsTest{
			NewChooseChannelsSingleTest(0.5, 0, "Low", false),
			NewChooseChannelsSingleTest(1.0, 0, "Low", false),
			NewChooseChannelsSingleTest(2.0, 0, "Low", false),
			NewChooseChannelsSingleTest(5.0, 0, "Low", false),
			NewChooseChannelsSingleTest(10.0, 0, "Low", false),
			NewChooseChannelsSingleTest(20.0, 0, "Low", false),
			NewChooseChannelsSingleTest(30.0, 0, "Low", false),
			NewChooseChannelsSingleTest(50.0, 0, "Low", false),
			NewChooseChannelsSingleTest(100.0, 0, "Standard", false),
			NewChooseChannelsSingleTest(300.0, 0, "Standard", false),
			NewChooseChannelsSingleTest(400.0, 0, "High", false),
			NewChooseChannelsSingleTest(1000.0, 0, "High", false),
			NewChooseChannelsSingleTest(1500.0, 0, "High", false),
			NewChooseChannelsTest([]float64{0.5, 50.0, 150.0, 500.0}, []int{0, 0, 0, 0}, []string{"Standard", "Standard", "Standard", "Standard"}, false),
			NewChooseChannelsTest([]float64{0.5, 0.0, 150.0, 500.0}, []int{0, 0, 0, 0}, []string{"Standard", "", "Standard", "Standard"}, false),
		},
	}).Run(t)
}

// TestChannelChoiceOverlapHigh overlapping tips with the same maximum is an error because we can't distinguish them at the max volume
func TestChannelChoiceOverlapHigh(t *testing.T) {
	idGen := id.NewIDGenerator("ChooseChannelHamilton")
	(&ChooseChannelsTests{
		Tips: []*wtype.LHTip{
			NewChooseChannelsSingleTestTip(idGen, "Low", 0.5, 50.0),
			NewChooseChannelsSingleTestTip(idGen, "Standard", 20.0, 1000.0),
			NewChooseChannelsSingleTestTip(idGen, "High", 500.0, 1000.0),
		},
		Heads: []*wtype.LHHeadAssembly{
			NewChooseChannelsSingleTestHead(idGen, 0, 0.5, 1000.0),
		},
		Tests: []*ChooseChannelsTest{
			NewChooseChannelsSingleTest(0.5, 0, "", true),
		},
	}).Run(t)
}

// TestChannelChoiceOverlapHigh overlapping tips with one contained by another
func TestChannelChoiceOverlapMid(t *testing.T) {
	idGen := id.NewIDGenerator("ChooseChannelHamilton")
	(&ChooseChannelsTests{
		Tips: []*wtype.LHTip{
			NewChooseChannelsSingleTestTip(idGen, "Wide", 0.5, 100.0),
			NewChooseChannelsSingleTestTip(idGen, "Narrow", 10.0, 50.0),
		},
		Heads: []*wtype.LHHeadAssembly{
			NewChooseChannelsSingleTestHead(idGen, 0, 0.5, 1000.0),
		},
		Tests: []*ChooseChannelsTest{
			NewChooseChannelsSingleTest(0.5, 0, "Wide", false),
			NewChooseChannelsSingleTest(9.9, 0, "Wide", false),
			NewChooseChannelsSingleTest(10.0, 0, "Narrow", false),
			NewChooseChannelsSingleTest(50.0, 0, "Narrow", false),
			NewChooseChannelsSingleTest(50.1, 0, "Wide", false),
			NewChooseChannelsSingleTest(100.0, 0, "Wide", false),
			NewChooseChannelsTest([]float64{50.0, 51.0, 51.0, 50.0}, []int{0, 0, 0, 0}, []string{"Wide", "Wide", "Wide", "Wide"}, false),
		},
	}).Run(t)
}

// TestChannelChoiceNoOverlap there's a gap between 50 and 100 ul
func TestChannelChoiceNoOverlap(t *testing.T) {
	idGen := id.NewIDGenerator("ChooseChannelHamilton")
	(&ChooseChannelsTests{
		Tips: []*wtype.LHTip{
			NewChooseChannelsSingleTestTip(idGen, "Low", 0.5, 50.0),
			NewChooseChannelsSingleTestTip(idGen, "High", 100.0, 1000.0),
		},
		Heads: []*wtype.LHHeadAssembly{
			NewChooseChannelsSingleTestHead(idGen, 0, 0.5, 1000.0),
		},
		Tests: []*ChooseChannelsTest{
			NewChooseChannelsSingleTest(0.5, 0, "Low", false),
			NewChooseChannelsSingleTest(1.0, 0, "Low", false),
			NewChooseChannelsSingleTest(2.0, 0, "Low", false),
			NewChooseChannelsSingleTest(5.0, 0, "Low", false),
			NewChooseChannelsSingleTest(10.0, 0, "Low", false),
			NewChooseChannelsSingleTest(20.0, 0, "Low", false),
			NewChooseChannelsSingleTest(30.0, 0, "Low", false),
			NewChooseChannelsSingleTest(50.0, 0, "Low", false),
			NewChooseChannelsSingleTest(75.0, 0, "Low", false), // between the two tip options
			NewChooseChannelsSingleTest(100.0, 0, "High", false),
			NewChooseChannelsSingleTest(300.0, 0, "High", false),
			NewChooseChannelsSingleTest(400.0, 0, "High", false),
			NewChooseChannelsSingleTest(1000.0, 0, "High", false),
			NewChooseChannelsSingleTest(1500.0, 0, "High", false),
		},
	}).Run(t)
}

// TestIdenticalTipsError any setup where there are multiple tips which can't be distinguished should result in an error
func TestIdenticalTipsError(t *testing.T) {
	idGen := id.NewIDGenerator("ChooseChannelMultipleTips")
	(&ChooseChannelsTests{
		Tips: []*wtype.LHTip{
			NewChooseChannelsSingleTestTip(idGen, "Low", 0.5, 20.0),
			NewChooseChannelsSingleTestTip(idGen, "AlsoLow", 0.5, 20.0),
		},
		Heads: []*wtype.LHHeadAssembly{
			NewChooseChannelsSingleTestHead(idGen, 0, 0.5, 20.0),
			NewChooseChannelsSingleTestHead(idGen, 1, 20.0, 250.0),
		},
		Tests: []*ChooseChannelsTest{
			NewChooseChannelsSingleTest(0.47, 0, "", true),
			NewChooseChannelsSingleTest(0.5, 0, "", true),
			NewChooseChannelsSingleTest(1.0, 0, "", true),
			NewChooseChannelsSingleTest(2.0, 0, "", true),
			NewChooseChannelsSingleTest(5.0, 0, "", true),
			NewChooseChannelsSingleTest(10.0, 0, "", true),
			NewChooseChannelsSingleTest(20.0, 0, "", true),
			NewChooseChannelsSingleTest(30.0, 1, "", true),
			NewChooseChannelsSingleTest(50.0, 1, "", true),
			NewChooseChannelsSingleTest(100.0, 1, "", true),
			NewChooseChannelsSingleTest(200.0, 1, "", true),
			NewChooseChannelsSingleTest(300.0, 1, "", true),
		},
	}).Run(t)
}
