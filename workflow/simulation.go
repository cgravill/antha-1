package workflow

import "github.com/antha-lang/antha/utils"

func EmptySimulation() Simulation {
	return Simulation{
		Elements: SimulatedElements{
			Types:     make(SimulatedElementTypes),
			Instances: make(SimulatedElementInstances),
		},
	}
}

type Simulation struct {
	// The SimulationId is an Id created by the act of simulation. Thus
	// a workflow that is simulated twice will have the same WorkflowId
	// but different SimulationIds.
	SimulationId BasicId          `json:"SimulationId,omitempty"`
	Version      string           `json:"Version"`
	Start        string           `json:"Start"`
	End          string           `json:"End"`
	InDir        string           `json:"InDir"`
	OutDir       string           `json:"OutDir"`
	Errors       utils.ErrorSlice `json:"Errors"`

	Elements SimulatedElements `json:"Elements,omitempty"`
}

type SimulatedElements struct {
	Types     SimulatedElementTypes     `json:"Types,omitempty"`
	Instances SimulatedElementInstances `json:"Instances,omitempty"`
}

type SimulatedElementTypes map[ElementTypeName]SimulatedElementType

type SimulatedElementType struct {
	ElementType
	GoSrcPath            string
	AnthaSrcPath         string
	InputsFieldTypes     map[string]string
	OutputsFieldTypes    map[string]string
	ParametersFieldTypes map[string]string
	DataFieldTypes       map[string]string
}

type SimulatedElementInstances map[ElementInstanceId]SimulatedElementInstance

type ElementInstanceId string

type SimulatedElementInstance struct {
	Name      ElementInstanceName `json:"Name"`
	TypeName  ElementTypeName     `json:"TypeName"`
	ParentId  ElementInstanceId   `json:"ParentId,omitempty"`
	Error     string              `json:"Error,omitempty"`
	StatePath string              `json:"StatePath"`
}
