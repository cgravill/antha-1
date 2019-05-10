package liquidhandling

import (
	"fmt"
	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"testing"
)

type promptTest struct {
	Name         string
	Instructions []*wtype.LHInstruction
	Chain        *wtype.IChain
}

var promptTests = []promptTest{
	makeFirstPromptTest(),
	makeSecondPromptTest(),
	makeThirdPromptTest(),
	makeFourthPromptTest(),
	makeFifthPromptTest(),
}

func makeComponents(names ...string) []*wtype.LHComponent {
	ret := make([]*wtype.LHComponent, 0, len(names))
	for _, name := range names {
		ret = append(ret, makeComponent(name))
	}

	return ret
}

func makeComponent(name string) *wtype.LHComponent {
	return &wtype.LHComponent{
		CName: name,
		ID:    name,
	}
}

var counter int = 0

func getID() string {
	id := fmt.Sprintf("ID-%d", counter)
	counter++
	return id
}

func mix(components ...*wtype.LHComponent) *wtype.LHInstruction {
	if len(components) < 2 {
		panic(fmt.Sprintf("Need at least 2 components to do mix, got %d", len(components)))
	}

	return &wtype.LHInstruction{
		ID:      getID(),
		Type:    wtype.LHIMIX,
		Inputs:  components[0 : len(components)-1],
		Outputs: []*wtype.LHComponent{components[len(components)-1]},
	}
}

func prompt(components ...*wtype.LHComponent) *wtype.LHInstruction {
	if len(components)%2 != 0 {
		panic(fmt.Sprintf("Need even number of components to do prompt, got %d", len(components)))
	}

	return &wtype.LHInstruction{
		ID:      getID(),
		Type:    wtype.LHIPRM,
		Inputs:  components[0 : len(components)/2],
		Outputs: components[len(components)/2:],
	}
}

// B = MIX(A) C = PROMPT(B) D = MIX(C)
func makeFirstPromptTest() promptTest {
	name := "Simple - three layers, mix-prompt-mix"
	cmps := makeComponents("A", "B", "C", "D")
	insx := []*wtype.LHInstruction{
		mix(cmps[0], cmps[1]), prompt(cmps[1], cmps[2]), mix(cmps[2], cmps[3]),
	}
	chain := wtype.MakeNewIChain([]*wtype.LHInstruction{insx[0]}, []*wtype.LHInstruction{insx[1]}, []*wtype.LHInstruction{insx[2]})

	return promptTest{
		Name:         name,
		Instructions: insx,
		Chain:        chain,
	}
}

// this test and the next jointly show that multi-component prompts behave identically with
// sets of single component prompts if they arrive at the same place in the chain
// MIX(A)->B MIX(C)->D PROMPT(B,D)->(E,F) MIX(E)->G MIX(F)->H
func makeSecondPromptTest() promptTest {
	name := "Multi-component prompts"

	cmps := makeComponents("A", "B", "C", "D", "E", "F", "G", "H")

	insx := []*wtype.LHInstruction{mix(cmps[0], cmps[1]), mix(cmps[2], cmps[3]), prompt(cmps[1], cmps[3], cmps[4], cmps[5]), mix(cmps[4], cmps[6]), mix(cmps[5], cmps[7])}

	// the middle prompt is what should result after aggregation since all prompts have the same (empty) message

	insgs := [][]*wtype.LHInstruction{{insx[0], insx[1]}, {insx[2]}, {insx[3], insx[4]}}

	chain := wtype.MakeNewIChain(insgs...)
	return promptTest{
		Name:         name,
		Instructions: insx,
		Chain:        chain,
	}
}

// MIX(A)->B MIX(C)->D PROMPT(B,D)->(E,F) MIX(E)->G MIX(F)->H
func makeThirdPromptTest() promptTest {
	name := "Aggregation test 1 : simple merger"

	cmps := makeComponents("A", "B", "C", "D", "E", "F", "G", "H")

	insx := []*wtype.LHInstruction{mix(cmps[0], cmps[1]), mix(cmps[2], cmps[3]), prompt(cmps[1], cmps[4]), prompt(cmps[3], cmps[5]), mix(cmps[4], cmps[6]), mix(cmps[5], cmps[7])}

	// the middle prompt is what should result after aggregation since all prompts have the same (empty) message

	insgs := [][]*wtype.LHInstruction{{insx[0], insx[1]}, {prompt(cmps[1], cmps[3], cmps[4], cmps[5])}, {insx[4], insx[5]}}

	chain := wtype.MakeNewIChain(insgs...)
	return promptTest{
		Name:         name,
		Instructions: insx,
		Chain:        chain,
	}
}

// these next two tests illustrate the difference in behaviour:
// in this case the two prompts end up separated by a mix because
// of the way in which instruction chains are built up
// prompt(A) -> B prompt(B) -> C MIX(C) -> D
// prompt(E) -> F MIX(F) ->G
func makeFourthPromptTest() promptTest {
	name := "Aggregation test 2 : prompts separated by mix"
	cmps := makeComponents("A", "B", "C", "D", "E", "F", "G")
	insx := []*wtype.LHInstruction{prompt(cmps[0], cmps[1]), prompt(cmps[1], cmps[2]), mix(cmps[2], cmps[3]), prompt(cmps[4], cmps[5]), mix(cmps[5], cmps[6])}
	insgs := [][]*wtype.LHInstruction{{prompt(cmps[0], cmps[4], cmps[1], cmps[5])}, {insx[4]}, {insx[1]}, {insx[2]}}

	chain := wtype.MakeNewIChain(insgs...)

	return promptTest{
		Name:         name,
		Instructions: insx,
		Chain:        chain,
	}

}

// in this instance the prompt cannot get split up, so it must end up after all its inputs
// are available
// prompt(A) -> B prompt(B,E) -> (C,F) MIX(C) -> D
// MIX(F) ->G
func makeFifthPromptTest() promptTest {
	name := "multi-component prompt - should occur at correct location in chain"
	cmps := makeComponents("A", "B", "C", "D", "E", "F", "G")

	insx := []*wtype.LHInstruction{prompt(cmps[0], cmps[1]), prompt(cmps[1], cmps[4], cmps[2], cmps[5]), mix(cmps[2], cmps[3]), mix(cmps[5], cmps[6])}
	insgs := [][]*wtype.LHInstruction{{insx[0]}, {insx[1]}, {insx[2], insx[3]}}

	chain := wtype.MakeNewIChain(insgs...)

	return promptTest{
		Name:         name,
		Instructions: insx,
		Chain:        chain,
	}
}

func chainsEqual(ch1, ch2 *wtype.IChain) bool {
	// panic if either chain is invalid

	if err := ch1.AssertInstructionsSeparate(); err != nil {
		panic(err)
	} else if err := ch2.AssertInstructionsSeparate(); err != nil {
		panic(err)
	}

	if ch1.Height() != ch2.Height() {
		return false
	}

	cur1 := ch1
	cur2 := ch2

	for {
		if cur1 == nil {
			break
		}

		if !valuesEqual(cur1.Values, cur2.Values) {
			return false
		}

		cur1 = cur1.Child
		cur2 = cur2.Child
	}

	return true
}

// we care that the same IDs come in and out in the same order
func summariseInstructionCmps(v []*wtype.LHInstruction) [][]string {
	sa := [][]string{}

	for _, ins := range v {
		saa := []string{}

		if ins.Type == wtype.LHIMIX {
			s := ""
			for _, c := range ins.Inputs {
				s += c.ID + ","
			}

			s += "-->" + ins.Outputs[0].ID

			saa = append(saa, s)
		} else if ins.Type == wtype.LHIPRM {
			for i := 0; i < len(ins.Inputs); i++ {
				saa = append(saa, ins.Inputs[i].ID+"-->"+ins.Outputs[i].ID)
			}
		}
		sa = append(sa, saa)
	}

	return sa
}

func cmpSummariesEqual(sa1, sa2 [][]string) bool {
	if len(sa1) != len(sa2) {
		return false
	}
	m := make(map[string]bool)

	for i := 0; i < len(sa1); i++ {
		for _, s := range sa1[i] {
			m[s] = true
		}

	}

	for i := 0; i < len(sa1); i++ {
		for _, s := range sa2[i] {
			if !m[s] {
				return false
			}
		}
	}

	return true
}

func valuesEqual(v1, v2 []*wtype.LHInstruction) bool {
	if len(v1) != len(v2) {
		return false
	}

	if len(v1) == 0 {
		return true
	}

	// all instruction types must be identical (otherwise AssertInstructionsSeparate would have failed)
	// so it is sufficient to compare the first

	if v1[0].Type != v2[0].Type {
		return false
	}

	// urgh

	cmps1 := summariseInstructionCmps(v1)
	cmps2 := summariseInstructionCmps(v2)

	return cmpSummariesEqual(cmps1, cmps2)
}

func asMap(inss []*wtype.LHInstruction) map[string]*wtype.LHInstruction {
	r := make(map[string]*wtype.LHInstruction, len(inss))

	for _, ins := range inss {
		r[ins.ID] = ins
	}

	return r
}

func TestPrompts(t *testing.T) {
	for _, pt := range promptTests {
		t.Run(pt.Name, func(t *testing.T) {
			ch, err := buildInstructionChain(asMap(pt.Instructions))

			if err != nil {
				t.Errorf("Got error building instruction chain: %v", err)
			}

			if !chainsEqual(pt.Chain, ch) {
				t.Errorf("returned iChain not as anticipated: expected %v got %v", pt.Chain, ch)

				ch.Print()
			}
		})
	}
}
