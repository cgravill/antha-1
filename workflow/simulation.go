package workflow

import "github.com/antha-lang/antha/utils"

func (wf *Workflow) NewSimulation() error {
	if simId, err := RandomBasicId(wf.WorkflowId); err != nil {
		return err
	} else {
		wf.Simulation = Simulation{
			SimulationId: simId,
			Elements:     make(SimulatedElements),
		}
		return nil
	}
}

type Simulation struct {
	// The SimulationId is an Id created by the act of simulation. Thus
	// a workflow that is simulated twice will have the same WorkflowId
	// but different SimulationIds.
	SimulationId BasicId `json:"SimulationId,omitempty"`
	Version      string  `json:"Version"`
	Start        string  `json:"Start"`
	End          string  `json:"End"`
	InDir        string  `json:"InDir"`
	OutDir       string  `json:"OutDir"`

	Elements SimulatedElements `json:"Elements,omitempty"`

	Errors utils.ErrorSlice `json:"Errors"`
}

type SimulatedElements map[ElementId]*SimulatedElement

type ElementId string

type SimulatedElement struct {
	ElementInstanceName ElementInstanceName `json:"ElementName"`
	ElementTypeName     ElementTypeName     `json:"ElementTypeName"`
	StatePath           string              `json:"StatePath"`
	ParentElementId     ElementId           `json:"ParentElementId,omitempty"`
	Error               string              `json:"Error,omitempty"`
}
