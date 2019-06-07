package tests

import (
	"fmt"
	"testing"

	"github.com/Synthace/antha/antha/anthalib/wtype"
	"github.com/Synthace/antha/codegen"
	"github.com/Synthace/antha/instructions"
	"github.com/Synthace/antha/laboratory/effects"
	"github.com/Synthace/antha/laboratory/testlab"
	"github.com/Synthace/antha/target"
	"github.com/Synthace/antha/target/human"
	"github.com/Synthace/antha/workflow"
)

type lowLevelTestInst struct {
	instructions.DependsMixin
	instructions.IdMixin
}

type highLevelTestInst struct{}

type testDriver struct{}

var testDriverSelector = instructions.NameValue{
	Name:  target.DriverSelectorV1Name,
	Value: "antha.test.v0",
}

func (a *testDriver) CanCompile(req instructions.Request) bool {
	can := instructions.Request{
		Selector: []instructions.NameValue{
			testDriverSelector,
		},
	}
	return can.Contains(req)
}

func (a *testDriver) Compile(labEffects *effects.LaboratoryEffects, dir string, cmds []instructions.Node) (instructions.Insts, error) {
	for _, n := range cmds {
		if c, ok := n.(*instructions.Command); !ok {
			return nil, fmt.Errorf("unexpected node %T", n)
		} else if _, ok := c.Inst.(*highLevelTestInst); !ok {
			return nil, fmt.Errorf("unexpected inst %T", c.Inst)
		}
	}
	return instructions.Insts{&lowLevelTestInst{}}, nil
}

func (a *testDriver) Connect(wf *workflow.Workflow) error {
	return nil
}

func (a *testDriver) Close() {}

func (a *testDriver) Id() workflow.DeviceInstanceID {
	return workflow.DeviceInstanceID("testDevice")
}

func TestWellFormed(t *testing.T) {
	labEffects := testlab.NewTestLabEffects(nil)

	nodes := make([]instructions.Node, 4)
	for idx := 0; idx < len(nodes); idx++ {
		m := &instructions.Command{
			Request: instructions.Request{
				Selector: []instructions.NameValue{
					target.DriverSelectorV1Mixer,
				},
			},
			Inst: &wtype.LHInstruction{},
			From: []instructions.Node{
				&instructions.UseComp{},
				&instructions.UseComp{},
				&instructions.UseComp{},
			},
		}
		u := &instructions.UseComp{
			From: []instructions.Node{m},
		}

		t := &instructions.Command{
			Request: instructions.Request{
				Selector: []instructions.NameValue{
					testDriverSelector,
				},
			},
			Inst: &highLevelTestInst{},
			From: []instructions.Node{u},
		}

		nodes[idx] = t
	}

	tgt := target.New()
	if err := tgt.AddDevice(&testDriver{}); err != nil {
		t.Fatal(err)
	}
	if err := human.New(labEffects.IDGenerator).DetermineRole(tgt); err != nil {
		t.Fatal(err)
	}

	if insts, err := codegen.Compile(labEffects, "", tgt, nodes); err != nil {
		t.Fatal(err)
	} else if l := len(insts); l == 0 {
		t.Errorf("expected > %d instructions found %d", 0, l)
	} else if last, ok := insts[l-1].(*lowLevelTestInst); !ok {
		t.Errorf("expected testInst found %T", insts[l-1])
	} else if n := len(last.Depends); n != 1 {
		t.Errorf("expected %d dependencies found %d", 1, n)
	}
}
