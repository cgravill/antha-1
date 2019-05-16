package execute

import (
	"fmt"
	"time"

	"github.com/antha-lang/antha/antha/anthalib/mixer"
	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/antha/anthalib/wunit"
	"github.com/antha-lang/antha/instructions"
	"github.com/antha-lang/antha/laboratory"
	"github.com/antha-lang/antha/target"
)

// SetInputPlate Indicate to the scheduler the the contents of the plate is user
// supplied. This modifies the argument to mark each well as such.
func SetInputPlate(lab *laboratory.Laboratory, plate *wtype.Plate) {
	lab.SampleTracker.SetInputPlate(lab.IDGenerator, plate)
}

// An IncubateOpt are options to an incubate command
type IncubateOpt struct {
	// Time for which to incubate component
	Time wunit.Time
	// Temperature at which to incubate component
	Temp wunit.Temperature
	// Rate at which to shake incubator (force is device dependent)
	ShakeRate wunit.Rate
	// Radius at which ShakeRate is defined
	ShakeRadius wunit.Length

	// Time for which to pre-heat incubator
	PreTemp wunit.Temperature
	// Temperature at which to pre-heat incubator
	PreTime wunit.Time
	// Rate at which to pre-heat incubator
	PreShakeRate wunit.Rate
	// Radius at which PreShakeRate is defined
	PreShakeRadius wunit.Length
}

func newCompFromComp(lab *laboratory.Laboratory, in *wtype.Liquid) *wtype.Liquid {
	comp := in.Dup(lab.IDGenerator)
	comp.ID = lab.IDGenerator.NextID()
	comp.BlockID = wtype.NewBlockID(string(lab.Workflow.Simulation.SimulationId))
	comp.SetGeneration(comp.Generation() + 1)

	lab.Maker.UpdateAfterInst(in.ID, comp.ID)
	lab.SampleTracker.UpdateIDOf(in.ID, comp.ID)

	return comp
}

// Incubate incubates a component
func Incubate(lab *laboratory.Laboratory, in *wtype.Liquid, opt IncubateOpt) *wtype.Liquid {
	// nolint: gosimple
	innerInst := &instructions.IncubateInst{
		Time:           opt.Time,
		Temp:           opt.Temp,
		ShakeRate:      opt.ShakeRate,
		ShakeRadius:    opt.ShakeRadius,
		PreTemp:        opt.PreTemp,
		PreTime:        opt.PreTime,
		PreShakeRate:   opt.PreShakeRate,
		PreShakeRadius: opt.PreShakeRadius,
	}

	inst := &instructions.CommandInst{
		Args:   []*wtype.Liquid{in},
		Result: []*wtype.Liquid{newCompFromComp(lab, in)},
		Command: &instructions.Command{
			Inst: innerInst,
			Request: instructions.Request{
				Selector: []instructions.NameValue{
					target.DriverSelectorV1ShakerIncubator,
				},
			},
		},
	}

	lab.Trace.Issue(inst)
	return inst.Result[0]
}

// prompt... works pretty much like Handle does
// but passes the instruction to the planner
// in future this should generate handles as side-effects

type mixerPromptOpts struct {
	Component   *wtype.Liquid
	ComponentIn *wtype.Liquid
	Message     string
	WaitTime    wunit.Time
}

// MixerPrompt prompts user with a message during mixer execution
func MixerPrompt(lab *laboratory.Laboratory, in *wtype.Liquid, message string) *wtype.Liquid {
	inst := mixerPrompt(lab,
		mixerPromptOpts{
			Component:   newCompFromComp(lab, in),
			ComponentIn: in,
			Message:     message,
		},
	)
	lab.Trace.Issue(inst)
	return inst.Result[0]
}

// MixerWait prompts user with a message during mixer execution and waits for the specifed time before resuming.
func MixerWait(lab *laboratory.Laboratory, in *wtype.Liquid, time wunit.Time, message string) *wtype.Liquid {
	inst := mixerPrompt(lab,
		mixerPromptOpts{
			Component:   newCompFromComp(lab, in),
			ComponentIn: in,
			Message:     message,
			WaitTime:    time,
		},
	)
	lab.Trace.Issue(inst)
	return inst.Result[0]
}

// ExecuteMixes will ensure that all mix activities
// in a workflow prior to this point must be completed before Mix instructions after this point.
func ExecuteMixes(lab *laboratory.Laboratory, liquid *wtype.LHComponent) *wtype.LHComponent {
	return MixerPrompt(lab, liquid, wtype.MAGICBARRIERPROMPTSTRING)
}

// Prompt prompts user with a message
func Prompt(lab *laboratory.Laboratory, in *wtype.Liquid, message string) *wtype.Liquid {
	inst := &instructions.CommandInst{
		Args:   []*wtype.Liquid{in},
		Result: []*wtype.Liquid{newCompFromComp(lab, in)},
		Command: &instructions.Command{
			Inst: &instructions.PromptInst{
				Message: message,
			},
			Request: instructions.Request{
				Selector: []instructions.NameValue{
					target.DriverSelectorV1Human,
				},
			},
		},
	}

	lab.Trace.Issue(inst)
	return inst.Result[0]
}

func mixerPrompt(lab *laboratory.Laboratory, opts mixerPromptOpts) *instructions.CommandInst {
	inst := wtype.NewLHPromptInstruction(lab.IDGenerator)
	inst.SetGeneration(opts.ComponentIn.Generation())
	inst.Message = opts.Message
	// precision will be cut to the nearest second
	inst.WaitTime = opts.WaitTime.AsDuration().Round(time.Second)
	inst.AddOutput(opts.Component)
	inst.AddInput(opts.ComponentIn)
	inst.PassThrough[opts.ComponentIn.ID] = opts.Component

	return &instructions.CommandInst{
		Args:   []*wtype.Liquid{opts.ComponentIn},
		Result: []*wtype.Liquid{opts.Component},
		Command: &instructions.Command{
			Inst: inst,
			Request: instructions.Request{
				Selector: []instructions.NameValue{
					target.DriverSelectorV1Prompter,
				},
			},
		},
	}
}

// PlateReadOpts defines plate-reader absorbance options
type PlateReadOpts struct {
	Sample  *wtype.Liquid
	Options string
}

func readPlate(lab *laboratory.Laboratory, opts PlateReadOpts) *instructions.CommandInst {
	inst := wtype.NewPRInstruction(lab.IDGenerator)
	inst.ComponentIn = opts.Sample

	// Clone the component to represent the result of the AbsorbanceRead
	inst.ComponentOut = newCompFromComp(lab, opts.Sample)
	inst.Options = opts.Options

	return &instructions.CommandInst{
		Args:   []*wtype.Liquid{opts.Sample},
		Result: []*wtype.Liquid{inst.ComponentOut},
		Command: &instructions.Command{
			Inst: inst,
			Request: instructions.Request{
				Selector: []instructions.NameValue{
					target.DriverSelectorV1WriteOnlyPlateReader,
				},
			},
		},
	}
}

// PlateRead reads absorbance of a component
func PlateRead(lab *laboratory.Laboratory, opt PlateReadOpts) *wtype.Liquid {
	inst := readPlate(lab, opt)
	lab.Trace.Issue(inst)
	return inst.Result[0]
}

// QPCROptions are the options for a QPCR request.
type QPCROptions struct {
	Reactions  []*wtype.Liquid
	Definition string
	Barcode    string
	TagAs      string
}

func runQPCR(lab *laboratory.Laboratory, opts QPCROptions, command string) *instructions.CommandInst {
	inst := &instructions.QPCRInstruction{ID: lab.IDGenerator.NextID()}
	inst.Command = command
	inst.ComponentIn = opts.Reactions
	inst.Definition = opts.Definition
	inst.Barcode = opts.Barcode
	inst.TagAs = opts.TagAs
	inst.ComponentOut = []*wtype.Liquid{}

	for _, r := range inst.ComponentIn {
		inst.ComponentOut = append(inst.ComponentOut, newCompFromComp(lab, r))
	}

	return &instructions.CommandInst{
		Args:   opts.Reactions,
		Result: inst.ComponentOut,
		Command: &instructions.Command{
			Inst: inst,
			Request: instructions.Request{
				Selector: []instructions.NameValue{
					target.DriverSelectorV1QPCRDevice,
				},
			},
		},
	}
}

// RunQPCRExperiment starts a new QPCR experiment, using an experiment input file.
func RunQPCRExperiment(lab *laboratory.Laboratory, opt QPCROptions) []*wtype.Liquid {
	inst := runQPCR(lab, opt, "RunExperiment")
	lab.Trace.Issue(inst)
	return inst.Result
}

// RunQPCRFromTemplate starts a new QPCR experiment, using a template input file.
func RunQPCRFromTemplate(lab *laboratory.Laboratory, opt QPCROptions) []*wtype.Liquid {
	inst := runQPCR(lab, opt, "RunExperimentFromTemplate")
	lab.Trace.Issue(inst)
	return inst.Result
}

// NewComponent returns a new component given a component type
func NewComponent(lab *laboratory.Laboratory, typ string) *wtype.Liquid {
	c, err := lab.Inventory.Components.NewComponent(typ)
	if err != nil {
		panic(fmt.Sprintf("cannot make component %s: %s", typ, err))
	}
	return c
}

// NewPlate returns a new plate given a plate type
func NewPlate(lab *laboratory.Laboratory, typ wtype.PlateTypeName) *wtype.Plate {
	p, err := lab.Inventory.Plates.NewPlate(typ)
	if err != nil {
		panic(fmt.Sprintf("cannot make plate %s: %s", typ, err))
	}
	return p
}

func mix(lab *laboratory.Laboratory, inst *wtype.LHInstruction) *instructions.CommandInst {
	inst.BlockID = wtype.NewBlockID(string(lab.Workflow.Simulation.SimulationId))
	inst.Outputs[0].BlockID = inst.BlockID
	result := inst.Outputs[0]
	//result.BlockID = inst.BlockID // DELETEME

	mx := 0
	// from the protocol POV components need to be passed by value
	for i, c := range wtype.CopyComponentArray(lab.IDGenerator, inst.Inputs) {
		if c.CName == "" {
			panic("Nameless Component used in Mix - this is not permitted")
		}
		c.Order = i

		//result.MixPreserveTvol(c)
		if c.Generation() > mx {
			mx = c.Generation()
		}
		lab.Maker.UpdateAfterInst(c.ID, result.ID)
	}

	inst.SetGeneration(mx)
	result.SetGeneration(mx + 1)
	result.DeclareInstance()

	return &instructions.CommandInst{
		Args: inst.Inputs,
		Command: &instructions.Command{
			Inst: inst,
			Request: instructions.Request{
				Selector: []instructions.NameValue{
					target.DriverSelectorV1Mixer,
				},
			},
		},
		Result: []*wtype.Liquid{result},
	}
}

func genericMix(lab *laboratory.Laboratory, generic *wtype.LHInstruction) *wtype.Liquid {
	inst := mix(lab, generic)
	lab.Trace.Issue(inst)
	if generic.Welladdress != "" {
		err := inst.Result[0].SetWellLocation(generic.Welladdress)
		if err != nil {
			panic(err)
		}
	}
	return inst.Result[0]
}

// Mix mixes components
func Mix(lab *laboratory.Laboratory, components ...*wtype.Liquid) *wtype.Liquid {
	return genericMix(lab, mixer.GenericMix(lab, mixer.MixOptions{
		Inputs: components,
	}))
}

// MixInto mixes components
func MixInto(lab *laboratory.Laboratory, outplate *wtype.Plate, address string, components ...*wtype.Liquid) *wtype.Liquid {
	return genericMix(lab, mixer.GenericMix(lab, mixer.MixOptions{
		Inputs:      components,
		Destination: outplate,
		Address:     address,
	}))
}

// MixNamed mixes components
func MixNamed(lab *laboratory.Laboratory, outplatetype wtype.PlateTypeName, address string, platename string, components ...*wtype.Liquid) *wtype.Liquid {
	return genericMix(lab, mixer.GenericMix(lab, mixer.MixOptions{
		Inputs:    components,
		PlateType: outplatetype,
		Address:   address,
		PlateName: platename,
	}))
}

// MixTo mixes components
//
// TODO: Addresses break dependence information. Deprecated.
func MixTo(lab *laboratory.Laboratory, outplatetype wtype.PlateTypeName, address string, platenum int, components ...*wtype.Liquid) *wtype.Liquid {
	return genericMix(lab, mixer.GenericMix(lab, mixer.MixOptions{
		Inputs:    components,
		PlateType: outplatetype,
		Address:   address,
		PlateNum:  platenum,
	}))
}

// SplitSample is essentially an inverse mix: takes one component and a volume and returns two
// the question is then over what happens subsequently.. unlike mix this does not have a
// destination as it's intrinsically a source operation
func SplitSample(lab *laboratory.Laboratory, component *wtype.Liquid, volume wunit.Volume) (removed, remaining *wtype.Liquid) {
	// at this point we cannot guarantee that volumes are accurate
	// so it's a case of best-efforts

	inst := splitSample(lab, component, volume)

	lab.Trace.Issue(inst)

	// protocol world must not be able to modify the copies seen here
	return inst.Result[0].Dup(lab.IDGenerator), inst.Result[1].Dup(lab.IDGenerator)
}

// Sample takes a sample of volume v from this liquid
func Sample(lab *laboratory.Laboratory, liquid *wtype.Liquid, v wunit.Volume) *wtype.Liquid {
	return mixer.Sample(lab, liquid, v)
}

func splitSample(lab *laboratory.Laboratory, component *wtype.Liquid, volume wunit.Volume) *instructions.CommandInst {

	split := wtype.NewLHSplitInstruction(lab.IDGenerator)

	// this will count as a mix-in-place effectively
	split.Inputs = append(split.Inputs, component.Dup(lab.IDGenerator))

	cmpMoving, cmpStaying := mixer.SplitSample(lab, component, volume)

	//the ID of the component that is staying has been updated
	lab.SampleTracker.UpdateIDOf(component.ID, cmpStaying.ID)

	split.AddOutput(cmpMoving)
	split.AddOutput(cmpStaying)

	// Create Instruction
	inst := &instructions.CommandInst{
		Args: []*wtype.Liquid{component},
		Command: &instructions.Command{
			Inst: split,
			Request: instructions.Request{
				Selector: []instructions.NameValue{
					target.DriverSelectorV1Mixer,
				},
			},
		},
		Result: []*wtype.Liquid{cmpMoving, cmpStaying},
	}

	return inst
}
