package target

import (
	"time"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/driver"
	"github.com/antha-lang/antha/instructions"
	"github.com/antha-lang/antha/laboratory/effects"
	"github.com/antha-lang/antha/laboratory/effects/id"
	"github.com/antha-lang/antha/microArch/driver/liquidhandling"
	lh "github.com/antha-lang/antha/microArch/scheduler/liquidhandling"
	"github.com/antha-lang/antha/utils"
)

// An Initializer is an instruction with initialization instructions
type Initializer interface {
	GetInitializers() []instructions.Inst
}

// A Finalizer is an instruction with finalization instructions
type Finalizer interface {
	GetFinalizers() []instructions.Inst
}

// A TimeEstimator is an instruction that can estimate its own execution time
type TimeEstimator interface {
	// GetTimeEstimate returns a time estimate for this instruction in seconds
	GetTimeEstimate() float64
}

// A TipEstimator is an instruction that uses tips and provides information on how many
type TipEstimator interface {
	// GetTipEstimates returns an estimate of how many tips this instruction will use
	GetTipEstimates() []wtype.TipEstimate
}

// Files are binary data encoded in tar.gz bytes
type Files struct {
	Type    string // Pseudo MIME-type describing contents of tarball
	Tarball []byte // Tar'ed and gzip'ed files
}

// An Order is a task to order physical components
type Order struct {
	Manual
	Mixes []*Mix
}

// A PlatePrep is a task to setup plates
type PlatePrep struct {
	Manual
	Mixes []*Mix
}

// A SetupMixer is a task to setup a mixer
type SetupMixer struct {
	Manual
	Mixes []*Mix
}

// A SetupIncubator is a task to setup an incubator
type SetupIncubator struct {
	Manual
	// Corresponding mix
	Mix              *Mix
	IncubationPlates []*wtype.Plate
}

var (
	_ TimeEstimator = (*Mix)(nil)
	_ Initializer   = (*Mix)(nil)
)

// A Mix is a task that runs a mixer
type Mix struct {
	instructions.DependsMixin
	instructions.IdMixin

	Device          effects.Device
	Request         *lh.LHRequest
	Properties      *liquidhandling.LHProperties
	FinalProperties *liquidhandling.LHProperties
	Final           map[string]string // Map from ids in Properties to FinalProperties
	Files           Files
	Initializers    []instructions.Inst
	Summary         *MixSummary
}

// GetTimeEstimate implements a TimeEstimator
func (a *Mix) GetTimeEstimate() float64 {
	est := 0.0

	if a.Request != nil {
		est = a.Request.TimeEstimate
	}

	return est
}

// GetTipEstimates implements a TipEstimator
func (a *Mix) GetTipEstimates() []wtype.TipEstimate {
	ret := []wtype.TipEstimate{}

	if a.Request != nil {
		ret = make([]wtype.TipEstimate, len(a.Request.TipsUsed))
		copy(ret, a.Request.TipsUsed)
	}

	return ret
}

// GetInitializers implements an Initializer
func (a *Mix) GetInitializers() []instructions.Inst {
	return a.Initializers
}

// MixSummary contains all the summary information which describes the mix instruction
type MixSummary struct {
	// Layout is a validated JSON summary of the deck layouts before and after
	// the mix operation takes place, suitable for consumption by the front end.
	// The JSON schema is available at microArch/driver/liquidhandling/schemas/layout.schema.json
	Layout []byte
	// Actions is a validated JSON summary of the steps taken during the liquidhandling
	// operation, suitable for consumption by the front end.
	// The JSON schema is available at microArch/driver/liquidhandling/schemas/actions.schema.json
	Actions []byte
}

// EnsureMixSummary constructs a new MixSummary object from the
// instructions and initial and final robot states if none already
// exists. An error is returned if the parameters are invalid or if
// either summary object fails JSON-schema validation
func (a *Mix) EnsureMixSummary(idGen *id.IDGenerator) error {
	if a.Summary != nil {
		return nil
	}
	layout, layoutErr := lh.SummarizeLayout(idGen, a.Properties, a.FinalProperties, a.Final)
	actions, actionsErr := lh.SummarizeActions(idGen, a.Properties, a.Request.InstructionTree)
	if err := (utils.ErrorSlice{layoutErr, actionsErr}.Pack()); err != nil {
		return err
	} else {
		a.Summary = &MixSummary{
			Layout:  layout,
			Actions: actions,
		}
		return nil
	}
}

// A Manual is human-aided interaction
type Manual struct {
	instructions.DependsMixin
	instructions.IdMixin

	Label   string
	Details string
}

var (
	_ Finalizer   = (*Run)(nil)
	_ Initializer = (*Run)(nil)
)

// Run calls on device
type Run struct {
	instructions.DependsMixin
	instructions.IdMixin

	Device  effects.Device
	Label   string
	Details string
	Calls   []driver.Call
	// Additional instructions to add to beginning of instruction stream.
	// Instructions are assumed to depend in FIFO order.
	Initializers []instructions.Inst
	// Additional instructions to add to end of instruction stream.
	// Instructions are assumed to depend in LIFO order.
	Finalizers []instructions.Inst
}

// GetInitializers implements an Initializer instruction
func (a *Run) GetInitializers() []instructions.Inst {
	return a.Initializers
}

// GetFinalizers implements a Finalizer instruction
func (a *Run) GetFinalizers() []instructions.Inst {
	return a.Finalizers
}

// Prompt is manual prompt instruction
type Prompt struct {
	instructions.DependsMixin
	instructions.IdMixin

	Message string
}

// Wait is a virtual instruction to hang dependencies on. A better name might
// been no-op.
type Wait struct {
	instructions.DependsMixin
	instructions.IdMixin
}

// TimedWait is a wait for a period of time.
type TimedWait struct {
	instructions.DependsMixin
	instructions.IdMixin

	Duration time.Duration
}
