package v1_2

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/microArch/driver/liquidhandling"
	"github.com/antha-lang/antha/workflow"
)

type workflowv1_2 struct {
	desc
	testOpt
	rawParams
	Version    string             `json:"version"`
	Properties workflowProperties `json:"Properties"`
}

type workflowProperties struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type rawParams struct {
	Parameters map[string]map[string]json.RawMessage `json:"Parameters"`
	Config     *opt                                  `json:"Config"`
}

type desc struct {
	Processes   map[string]process `json:"Processes"`
	Connections []connection       `json:"connections"`
}

type connection struct {
	Src port `json:"source"`
	Tgt port `json:"target"`
}

type process struct {
	Component string         `json:"component"`
	Metadata  screenPosition `json:"metadata"`
}

type screenPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type port struct {
	Process string `json:"process"`
	Port    string `json:"port"`
}

type testOpt struct {
	ComparisonOptions   string
	CompareInstructions bool
	CompareOutputs      bool
	Results             testResults
}

type testResults struct {
	MixTaskResults []mixTaskResult
}

type mixTaskResult struct {
	Instructions liquidhandling.SetOfRobotInstructions
	Outputs      map[string]*wtype.Plate
	TimeEstimate time.Duration
}

func (wf *workflowv1_2) MigratedElementParameters(name string) (*workflow.ElementParameterSet, error) {
	v, pr := wf.Parameters[name]
	if !pr {
		return nil, errors.New("parameters not present for element" + name)
	}
	pset := workflow.ElementParameterSet(make(map[workflow.ElementParameterName]json.RawMessage))

	for pname, pval := range v {
		epname := workflow.ElementParameterName(pname)
		pset[epname] = pval
	}
	return &pset, nil
}

func (wf *workflowv1_2) MigrateElement(name string) (*workflow.ElementInstance, error) {
	ei := &workflow.ElementInstance{}
	meta := &json.RawMessage{}

	v, pr := wf.Processes[name]
	if !pr {
		return nil, errors.New("element instance " + name + " not present")
	}
	ei.ElementTypeName = workflow.ElementTypeName(v.Component)
	if enc, err := json.Marshal(v.Metadata); err != nil {
		return nil, err
	} else if err := meta.UnmarshalJSON(enc); err != nil {
		return nil, err
	} else if params, err := wf.MigratedElementParameters(name); err != nil {
		return nil, err
	} else {
		ei.Meta = *meta
		ei.Parameters = *params
	}
	return ei, nil
}
