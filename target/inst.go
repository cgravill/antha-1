package target

import (
	"time"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/driver"
	"github.com/antha-lang/antha/laboratory/effects"
	"github.com/antha-lang/antha/laboratory/effects/id"
	"github.com/antha-lang/antha/microArch/driver/liquidhandling"
	lh "github.com/antha-lang/antha/microArch/scheduler/liquidhandling"
)

// An Initializer is an instruction with initialization instructions
type Initializer interface {
	GetInitializers() []effects.Inst
}

// A Finalizer is an instruction with finalization instructions
type Finalizer interface {
	GetFinalizers() []effects.Inst
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
	effects.DependsMixin
	effects.IdMixin
	effects.DeviceMixin

	Request         *lh.LHRequest
	Properties      *liquidhandling.LHProperties
	FinalProperties *liquidhandling.LHProperties
	Final           map[string]string // Map from ids in Properties to FinalProperties
	Files           Files
	Initializers    []effects.Inst
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
func (a *Mix) GetInitializers() []effects.Inst {
	return a.Initializers
}

// SummarizeLayout helper function to get a validated JSON summary of the deck layouts before and after
// the mix operation takes place, suitable for consumption by the front end.
// The JSON schema is available at microArch/driver/liquidhandling/schemas/layout.schema.json
func (a *Mix) SummarizeLayout(idGen *id.IDGenerator) ([]byte, error) {
	// Mix is the first time all of these things are tied together in a coherent way
	// so a helper here simplifies the API for callers, however the actual implementation of
	// SummarizeLayout makes more sense down in microArch/scheduler/liquidhandling
	return lh.SummarizeLayout(idGen, a.Properties, a.FinalProperties, a.Final)
}

// SummarizeActions helper function get a validated JSON summary of the steps taken during the liquidhandling
// operation, suitable for consumption by the front end.
// The JSON schema is available at microArch/driver/liquidhandling/schemas/actions.schema.json
func (a *Mix) SummarizeActions(idGen *id.IDGenerator) ([]byte, error) {
	// Mix is the first time all of these things are tied together in a coherent way
	// so a helper here simplifies the API for callers, however the actual implementation of
	// SummarizeLayout makes more sense down in microArch/scheduler/liquidhandling
	return lh.SummarizeActions(idGen, a.Properties, a.Request.InstructionTree)
}

// A Manual is human-aided interaction
type Manual struct {
	effects.DependsMixin
	effects.IdMixin
	effects.DeviceMixin

	Label   string
	Details string
}

var (
	_ Finalizer   = (*Run)(nil)
	_ Initializer = (*Run)(nil)
)

// Run calls on device
type Run struct {
	effects.DependsMixin
	effects.IdMixin
	effects.DeviceMixin

	Label   string
	Details string
	Calls   []driver.Call
	// Additional instructions to add to beginning of instruction stream.
	// Instructions are assumed to depend in FIFO order.
	Initializers []effects.Inst
	// Additional instructions to add to end of instruction stream.
	// Instructions are assumed to depend in LIFO order.
	Finalizers []effects.Inst
}

// GetInitializers implements an Initializer instruction
func (a *Run) GetInitializers() []effects.Inst {
	return a.Initializers
}

// GetFinalizers implements a Finalizer instruction
func (a *Run) GetFinalizers() []effects.Inst {
	return a.Finalizers
}

// Prompt is manual prompt instruction
type Prompt struct {
	effects.DependsMixin
	effects.IdMixin
	effects.NoDeviceMixin

	Message string
}

// Wait is a virtual instruction to hang dependencies on. A better name might
// been no-op.
type Wait struct {
	effects.DependsMixin
	effects.IdMixin
	effects.NoDeviceMixin
}

// TimedWait is a wait for a period of time.
type TimedWait struct {
	effects.DependsMixin
	effects.IdMixin
	effects.NoDeviceMixin

	Duration time.Duration
}
