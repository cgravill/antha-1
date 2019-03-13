package v1_2

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/antha-lang/antha/antha/anthalib/wtype"
	"github.com/antha-lang/antha/microArch/driver/liquidhandling"
	"github.com/antha-lang/antha/workflow"
)

type Workflow struct {
	Desc
	TestOpt
	RawParams
	Version    string             `json:"version"`
	Properties workflowProperties `json:"Properties"`
}

type workflowProperties struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RawParams struct {
	Parameters map[string]map[string]json.RawMessage `json:"Parameters"`
	Config     *Opt                                  `json:"Config"`
}

type Desc struct {
	Processes   map[string]Process `json:"Processes"`
	Connections []Connection       `json:"connections"`
}

type Connection struct {
	Src Port `json:"source"`
	Tgt Port `json:"target"`
}

type Process struct {
	Component string         `json:"component"`
	Metadata  screenPosition `json:"metadata"`
}

type screenPosition struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type Port struct {
	Process string `json:"process"`
	Port    string `json:"port"`
}

type TestOpt struct {
	ComparisonOptions   string
	CompareInstructions bool
	CompareOutputs      bool
	Results             TestResults
}

type TestResults struct {
	MixTaskResults []MixTaskResult
}

type MixTaskResult struct {
	Instructions liquidhandling.SetOfRobotInstructions
	Outputs      map[string]*wtype.Plate
	TimeEstimate time.Duration
}

func (wf *Workflow) MigratedElementParameters(name string) (*workflow.ElementParameterSet, error) {
	if v, pr := wf.Parameters[name]; !pr {
		return nil, errors.New("parameters not present for element" + name)
	} else {
		pset := workflow.ElementParameterSet(make(map[workflow.ElementParameterName]json.RawMessage))

		for pname, pval := range v {
			epname := workflow.ElementParameterName(pname)
			pset[epname] = pval
		}
		return &pset, nil
	}
}

func (wf *Workflow) MigratedElement(name string) (*workflow.ElementInstance, error) {
	ei := &workflow.ElementInstance{}
	meta := &json.RawMessage{}

	if v, pr := wf.Processes[name]; !pr {
		return nil, errors.New("element instance " + name + " not present")
	} else {
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
}
