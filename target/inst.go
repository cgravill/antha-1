package target

import (
	"time"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/driver"
	"github.com/antha-lang/antha/laboratory/effects"
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

type dependsMixin struct {
	Depends []effects.Inst
}

// DependsOn implements an Inst
func (a *dependsMixin) DependsOn() []effects.Inst {
	return a.Depends
}

// SetDependsOn implements an Inst
func (a *dependsMixin) SetDependsOn(x ...effects.Inst) {
	a.Depends = x
}

// AppendDependsOn implements an Inst
func (a *dependsMixin) AppendDependsOn(x ...effects.Inst) {
	a.Depends = append(a.Depends, x...)
}

type noDeviceMixin struct{}

// Device implements an Inst
func (a noDeviceMixin) Device() effects.Device {
	return nil
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
	dependsMixin

	Dev             effects.Device
	Request         *lh.LHRequest
	Properties      *liquidhandling.LHProperties
	FinalProperties *liquidhandling.LHProperties
	Final           map[string]string // Map from ids in Properties to FinalProperties
	Files           Files
	Initializers    []effects.Inst
}

// Device implements an Inst
func (a *Mix) Device() effects.Device {
	return a.Dev
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

// A Manual is human-aided interaction
type Manual struct {
	dependsMixin

	Dev     effects.Device
	Label   string
	Details string
}

// Device implements an Inst
func (a *Manual) Device() effects.Device {
	return a.Dev
}

var (
	_ Finalizer   = (*Run)(nil)
	_ Initializer = (*Run)(nil)
)

// Run calls on device
type Run struct {
	dependsMixin

	Dev     effects.Device
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

// Device implements an Inst
func (a *Run) Device() effects.Device {
	return a.Dev
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
	dependsMixin
	noDeviceMixin

	Message string
}

// Wait is a virtual instruction to hang dependencies on. A better name might
// been no-op.
type Wait struct {
	noDeviceMixin
	dependsMixin
}

// TimedWait is a wait for a period of time.
type TimedWait struct {
	dependsMixin
	noDeviceMixin

	Duration time.Duration
}
